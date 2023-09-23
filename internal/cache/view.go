// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cache implements the caching layer for gopls.
package cache

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/walteh/retab/internal/source"

	"github.com/walteh/retab/gen/gopls/span"
)

var _ source.View = (*View)(nil)

type View struct {
	id string

	// baseCtx is the context handed to NewView. This is the parent of all
	// background contexts created for this view.
	baseCtx context.Context

	// name is the user-specified name of this view.
	name string

	// lastOptions holds the most recent options on this view, used for detecting
	// major changes.
	//
	// Guarded by Session.viewMu.
	lastOptions *source.Options

	// Workspace information. The fields below are immutable, and together with
	// options define the build list. Any change to these fields results in a new
	// View.
	workspaceInformation // Go environment information

	// parseCache holds an LRU cache of recently parsed files.
	parseCache *parseCache

	// fs is the file source used to populate this view.
	fs *overlayFS

	// knownFiles tracks files that the view has accessed.
	// TODO(golang/go#57558): this notion is fundamentally problematic, and
	// should be removed.
	knownFilesMu sync.Mutex
	knownFiles   map[span.URI]bool

	// initCancelFirstAttempt can be used to terminate the view's first
	// attempt at initialization.
	initCancelFirstAttempt context.CancelFunc

	// Track the latest snapshot via the snapshot field, guarded by snapshotMu.
	//
	// Invariant: whenever the snapshot field is overwritten, destroy(snapshot)
	// is called on the previous (overwritten) snapshot while snapshotMu is held,
	// incrementing snapshotWG. During shutdown the final snapshot is
	// overwritten with nil and destroyed, guaranteeing that all observed
	// snapshots have been destroyed via the destroy method, and snapshotWG may
	// be waited upon to let these destroy operations complete.
	snapshotMu      sync.Mutex
	snapshot        *snapshot      // latest snapshot; nil after shutdown has been called
	releaseSnapshot func()         // called when snapshot is no longer needed
	snapshotWG      sync.WaitGroup // refcount for pending destroy operations

	// initialWorkspaceLoad is closed when the first workspace initialization has
	// completed. If we failed to load, we only retry if the go.mod file changes,
	// to avoid too many go/packages calls.
	initialWorkspaceLoad chan struct{}

	// initializationSema is used limit concurrent initialization of snapshots in
	// the view. We use a channel instead of a mutex to avoid blocking when a
	// context is canceled.
	//
	// This field (along with snapshot.initialized) guards against duplicate
	// initialization of snapshots. Do not change it without adjusting snapshot
	// accordingly.
	initializationSema chan struct{}
}

// workspaceInformation holds the defining features of the View workspace.
//
// This type is compared to see if the View needs to be reconstructed.
type workspaceInformation struct {
	// folder is the LSP workspace folder.
	folder span.URI

	// `go env` variables that need to be tracked by gopls.
	// goEnv

	// gomod holds the relevant go.mod file for this workspace.
	// gomod span.URI

	// The Go version in use: X in Go 1.X.
	// goversion int

	// The complete output of the go version command.
	// (Call gocommand.ParseGoVersionOutput to extract a version
	// substring such as go1.19.1 or go1.20-rc.1, go1.21-abcdef01.)
	// goversionOutput string

	// hasGopackagesDriver is true if the user has a value set for the
	// GOPACKAGESDRIVER environment variable or a gopackagesdriver binary on
	// their machine.
	// hasGopackagesDriver bool

	// inGOPATH reports whether the workspace directory is contained in a GOPATH
	// directory.
	// inGOPATH bool

	// goCommandDir is the dir to use for running go commands.
	//
	// The only case where this should matter is if we've narrowed the workspace to
	// a single nested module. In that case, the go command won't be able to find
	// the module unless we tell it the nested directory.
	// goCommandDir span.URI
}

// A ViewType describes how we load package information for a view.
//
// This is used for constructing the go/packages.Load query, and for
// interpreting missing packages, imports, or errors.
//
// Each view has a ViewType which is derived from its immutable workspace
// information -- any environment change that would affect the view type
// results in a new view.
type ViewType int

const (
	// GoPackagesDriverView is a view with a non-empty GOPACKAGESDRIVER
	// environment variable.
	GoPackagesDriverView ViewType = iota

	// GOPATHView is a view in GOPATH mode.
	//
	// I.e. in GOPATH, with GO111MODULE=off, or GO111MODULE=auto with no
	// go.mod file.
	GOPATHView

	// GoModuleView is a view in module mode with a single Go module.
	GoModuleView

	// GoWorkView is a view in module mode with a go.work file.
	GoWorkView

	// An AdHocView is a collection of files in a given directory, not in GOPATH
	// or a module.
	AdHocView
)

// ViewType derives the type of the view from its workspace information.
//
// TODO(rfindley): this logic is overlapping and slightly inconsistent with
// validBuildConfiguration. As part of zero-config-gopls (golang/go#57979), fix
// this inconsistency and consolidate on the ViewType abstraction.
func (w workspaceInformation) ViewType() ViewType {
	return AdHocView
}

// moduleMode reports whether the current snapshot uses Go modules.
//
// From https://go.dev/ref/mod, module mode is active if either of the
// following hold:
//   - GO111MODULE=on
//   - GO111MODULE=auto and we are inside a module or have a GOWORK value.
//
// Additionally, this method returns false if GOPACKAGESDRIVER is set.
//
// TODO(rfindley): use this more widely.
func (w workspaceInformation) moduleMode() bool {
	switch w.ViewType() {
	case GoModuleView, GoWorkView:
		return true
	default:
		return false
	}
}

// workspaceMode holds various flags defining how the gopls workspace should
// behave. They may be derived from the environment, user configuration, or
// depend on the Go version.
//
// TODO(rfindley): remove workspace mode, in favor of explicit checks.
type workspaceMode int

const (
	moduleMode workspaceMode = 1 << iota

	// tempModfile indicates whether or not the -modfile flag should be used.
	tempModfile
)

func (v *View) ID() string { return v.id }

// Name returns the user visible name of this view.
func (v *View) Name() string {
	return v.name
}

// Folder returns the folder at the base of this view.
func (v *View) Folder() span.URI {
	return v.folder
}

func minorOptionsChange(a, b *source.Options) bool {
	// TODO(rfindley): this function detects whether a view should be recreated,
	// but this is also checked by the getWorkspaceInformation logic.
	//
	// We should eliminate this redundancy.
	//
	// Additionally, this function has existed for a long time, but git history
	// suggests that it was added arbitrarily, not due to an actual performance
	// problem.
	//
	// Especially now that we have optimized reinitialization of the session, we
	// should consider just always creating a new view on any options change.

	// Check if any of the settings that modify our understanding of files have
	// been changed.
	if !reflect.DeepEqual(a.Env, b.Env) {
		return false
	}
	if !reflect.DeepEqual(a.DirectoryFilters, b.DirectoryFilters) {
		return false
	}
	if !reflect.DeepEqual(a.StandaloneTags, b.StandaloneTags) {
		return false
	}
	if a.ExpandWorkspaceToModule != b.ExpandWorkspaceToModule {
		return false
	}
	if a.MemoryMode != b.MemoryMode {
		return false
	}
	aBuildFlags := make([]string, len(a.BuildFlags))
	bBuildFlags := make([]string, len(b.BuildFlags))
	copy(aBuildFlags, a.BuildFlags)
	copy(bBuildFlags, b.BuildFlags)
	sort.Strings(aBuildFlags)
	sort.Strings(bBuildFlags)
	// the rest of the options are benign
	return reflect.DeepEqual(aBuildFlags, bBuildFlags)
}

// separated out from its sole use in locateTemplateFiles for testability
func fileHasExtension(path string, suffixes []string) bool {
	ext := filepath.Ext(path)
	if ext != "" && ext[0] == '.' {
		ext = ext[1:]
	}
	for _, s := range suffixes {
		if s != "" && ext == s {
			return true
		}
	}
	return false
}

const fileLimit = 100_000

var errExhausted = errors.New("exhausted")

func (s *snapshot) contains(uri span.URI) bool {

	inFolder := source.InDir(s.view.folder.Filename(), uri.Filename())

	if !inFolder {
		return false
	}

	return true
}

func (v *View) markKnown(uri span.URI) {
	v.knownFilesMu.Lock()
	defer v.knownFilesMu.Unlock()
	if v.knownFiles == nil {
		v.knownFiles = make(map[span.URI]bool)
	}
	v.knownFiles[uri] = true
}

// knownFile reports whether the specified valid URI (or an alias) is known to the view.
func (v *View) knownFile(uri span.URI) bool {
	v.knownFilesMu.Lock()
	defer v.knownFilesMu.Unlock()
	return v.knownFiles[uri]
}

// shutdown releases resources associated with the view, and waits for ongoing
// work to complete.
func (v *View) shutdown() {
	// Cancel the initial workspace load if it is still running.
	v.initCancelFirstAttempt()

	v.snapshotMu.Lock()
	if v.snapshot != nil {
		v.snapshot.cancel()
		v.releaseSnapshot()
		v.destroy(v.snapshot, "View.shutdown")
		v.snapshot = nil
		v.releaseSnapshot = nil
	}
	v.snapshotMu.Unlock()

	v.snapshotWG.Wait()
}

func (v *View) relevantChange(c source.FileModification) bool {
	// If the file is known to the view, the change is relevant.
	if v.knownFile(c.URI) {
		return true
	}

	// Note: CL 219202 filtered out on-disk changes here that were not known to
	// the view, but this introduces a race when changes arrive before the view
	// is initialized (and therefore, before it knows about files). Since that CL
	// had neither test nor associated issue, and cited only emacs behavior, this
	// logic was deleted.

	snapshot, release, err := v.getSnapshot()
	if err != nil {
		return false // view was shut down
	}
	defer release()
	return snapshot.contains(c.URI)
}

// While go list ./... skips directories starting with '.', '_', or 'testdata',
// gopls may still load them via file queries. Explicitly filter them out.
func (s *snapshot) IgnoredFile(uri span.URI) bool {
	// Fast path: if uri doesn't contain '.', '_', or 'testdata', it is not
	// possible that it is ignored.
	{
		uriStr := string(uri)
		if !strings.Contains(uriStr, ".") && !strings.Contains(uriStr, "_") && !strings.Contains(uriStr, "testdata") {
			return false
		}
	}

	return s.ignoreFilter.ignored(uri.Filename())
}

// An ignoreFilter implements go list's exclusion rules via its 'ignored' method.
type ignoreFilter struct {
	prefixes []string // root dirs, ending in filepath.Separator
}

// newIgnoreFilter returns a new ignoreFilter implementing exclusion rules
// relative to the provided directories.
func newIgnoreFilter(dirs []string) *ignoreFilter {
	f := new(ignoreFilter)
	for _, d := range dirs {
		f.prefixes = append(f.prefixes, filepath.Clean(d)+string(filepath.Separator))
	}
	return f
}

func (f *ignoreFilter) ignored(filename string) bool {
	for _, prefix := range f.prefixes {
		if suffix := strings.TrimPrefix(filename, prefix); suffix != filename {
			if checkIgnored(suffix) {
				return true
			}
		}
	}
	return false
}

// checkIgnored implements go list's exclusion rules.
// Quoting “go help list”:
//
//	Directory and file names that begin with "." or "_" are ignored
//	by the go tool, as are directories named "testdata".
func checkIgnored(suffix string) bool {
	// Note: this could be further optimized by writing a HasSegment helper, a
	// segment-boundary respecting variant of strings.Contains.
	for _, component := range strings.Split(suffix, string(filepath.Separator)) {
		if len(component) == 0 {
			continue
		}
		if component[0] == '.' || component[0] == '_' || component == "testdata" {
			return true
		}
	}
	return false
}

func (v *View) Snapshot() (source.Snapshot, func(), error) {
	return v.getSnapshot()
}

func (v *View) getSnapshot() (*snapshot, func(), error) {
	v.snapshotMu.Lock()
	defer v.snapshotMu.Unlock()
	if v.snapshot == nil {
		return nil, nil, errors.New("view is shutdown")
	}
	return v.snapshot, v.snapshot.Acquire(), nil
}

func (s *snapshot) initialize(ctx context.Context, firstAttempt bool) {
	select {
	case <-ctx.Done():
		return
	case s.view.initializationSema <- struct{}{}:
	}

	defer func() {
		<-s.view.initializationSema
	}()

	s.mu.Lock()
	initialized := s.initialized
	s.mu.Unlock()

	if initialized {
		return
	}

}

func (s *Session) getWorkspaceInformation(ctx context.Context, folder span.URI, options *source.Options) (workspaceInformation, error) {
	if err := checkPathCase(folder.Filename()); err != nil {
		return workspaceInformation{}, fmt.Errorf("invalid workspace folder path: %w; check that the casing of the configured workspace folder path agrees with the casing reported by the operating system", err)
	}
	info := workspaceInformation{
		folder: folder,
	}

	return info, nil
}

// findRootPattern looks for files with the given basename in dir or any parent
// directory of dir, using the provided FileSource. It returns the first match,
// starting from dir and search parents.
//
// The resulting string is either the file path of a matching file with the
// given basename, or "" if none was found.
func findRootPattern(ctx context.Context, dir, basename string, fs source.FileSource) (string, error) {
	for dir != "" {
		target := filepath.Join(dir, basename)
		fh, err := fs.ReadFile(ctx, span.URIFromPath(target))
		if err != nil {
			return "", err // context cancelled
		}
		if fileExists(fh) {
			return target, nil
		}
		// Trailing separators must be trimmed, otherwise filepath.Split is a noop.
		next, _ := filepath.Split(strings.TrimRight(dir, string(filepath.Separator)))
		if next == dir {
			break
		}
		dir = next
	}
	return "", nil
}

// OS-specific path case check, for case-insensitive filesystems.
var checkPathCase = defaultCheckPathCase

func defaultCheckPathCase(path string) error {
	return nil
}

const maxGovulncheckResultAge = 1 * time.Hour // Invalidate results older than this limit.
var timeNow = time.Now                        // for testing

// Copied from
// https://cs.opensource.google/go/go/+/master:src/cmd/go/internal/str/path.go;l=58;drc=2910c5b4a01a573ebc97744890a07c1a3122c67a
func globsMatchPath(globs, target string) bool {
	for globs != "" {
		// Extract next non-empty glob in comma-separated list.
		var glob string
		if i := strings.Index(globs, ","); i >= 0 {
			glob, globs = globs[:i], globs[i+1:]
		} else {
			glob, globs = globs, ""
		}
		if glob == "" {
			continue
		}

		// A glob with N+1 path elements (N slashes) needs to be matched
		// against the first N+1 path elements of target,
		// which end just before the N+1'th slash.
		n := strings.Count(glob, "/")
		prefix := target
		// Walk target, counting slashes, truncating at the N+1'th slash.
		for i := 0; i < len(target); i++ {
			if target[i] == '/' {
				if n == 0 {
					prefix = target[:i]
					break
				}
				n--
			}
		}
		if n > 0 {
			// Not enough prefix elements.
			continue
		}
		matched, _ := path.Match(glob, prefix)
		if matched {
			return true
		}
	}
	return false
}

var modFlagRegexp = regexp.MustCompile(`-mod[ =](\w+)`)
