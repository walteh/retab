// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cache implements the caching layer for gopls.
package cache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/walteh/retab/gen/gopls/event"
	// "github.com/walteh/retab/gen/gopls/gocommand"
	// "github.com/walteh/retab/gen/gopls/imports"
	"github.com/walteh/retab/gen/gopls/protocol"
	"github.com/walteh/retab/gen/gopls/span"

	// "github.com/walteh/retab/gen/gopls/vulncheck"
	"github.com/walteh/retab/gen/gopls/xcontext"
	"github.com/walteh/retab/internal/source"
	exec "golang.org/x/sys/execabs"
)

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

	// moduleUpgrades tracks known upgrades for module paths in each modfile.
	// Each modfile has a map of module name to upgrade version.
	moduleUpgradesMu sync.Mutex
	moduleUpgrades   map[span.URI]map[string]string

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
}

// effectiveGO111MODULE reports the value of GO111MODULE effective in the go
// command at this go version, assuming at least Go 1.16.
func (w workspaceInformation) effectiveGO111MODULE() go111module {
	switch w.GO111MODULE() {
	case "off":
		return off
	case "on", "":
		return on
	default:
		return auto
	}
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
	// An AdHocView is a collection of files in a given directory, not in GOPATH
	// or a module.

	AdHocView ViewType = iota
)

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

// SetFolderOptions updates the options of each View associated with the folder
// of the given URI.
//
// Calling this may cause each related view to be invalidated and a replacement
// view added to the session.
func (s *Session) SetFolderOptions(ctx context.Context, uri span.URI, options *source.Options) error {
	s.viewMu.Lock()
	defer s.viewMu.Unlock()

	for _, v := range s.views {
		if v.folder == uri {
			if err := s.setViewOptions(ctx, v, options); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Session) setViewOptions(ctx context.Context, v *View, options *source.Options) error {
	// no need to rebuild the view if the options were not materially changed
	if minorOptionsChange(v.lastOptions, options) {
		_, release := v.invalidateContent(ctx, nil, options, false)
		release()
		v.lastOptions = options
		return nil
	}
	return s.updateViewLocked(ctx, v, options)
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

// locateTemplateFiles ensures that the snapshot has mapped template files
// within the workspace folder.
func (s *snapshot) locateTemplateFiles(ctx context.Context) {
	if len(s.options.TemplateExtensions) == 0 {
		return
	}
	suffixes := s.options.TemplateExtensions

	searched := 0
	filterFunc := s.filterFunc()
	err := filepath.WalkDir(s.view.folder.Filename(), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if fileLimit > 0 && searched > fileLimit {
			return errExhausted
		}
		searched++
		if !fileHasExtension(path, suffixes) {
			return nil
		}
		uri := span.URIFromPath(path)
		if filterFunc(uri) {
			return nil
		}
		// Get the file in order to include it in the snapshot.
		// TODO(golang/go#57558): it is fundamentally broken to track files in this
		// way; we may lose them if configuration or layout changes cause a view to
		// be recreated.
		//
		// Furthermore, this operation must ignore errors, including context
		// cancellation, or risk leaving the snapshot in an undefined state.
		s.ReadFile(ctx, uri)
		return nil
	})
	if err != nil {
		event.Error(ctx, "searching for template files failed", err)
	}
}

func (s *snapshot) contains(uri span.URI) bool {
	return source.InDir(s.view.folder.Filename(), uri.Filename())
}

// filterFunc returns a func that reports whether uri is filtered by the currently configured
// directoryFilters.
func (s *snapshot) filterFunc() func(span.URI) bool {
	filterer := buildFilterer(s.view.folder.Filename(), s.view.gomodcache, s.options)
	return func(uri span.URI) bool {
		// Only filter relative to the configured root directory.
		if source.InDir(s.view.folder.Filename(), uri.Filename()) {
			return pathExcludedByFilter(strings.TrimPrefix(uri.Filename(), s.view.folder.Filename()), filterer)
		}
		return false
	}
}

func (v *View) relevantChange(c source.FileModification) bool {
	// If the file is known to the view, the change is relevant.
	if v.knownFile(c.URI) {
		return true
	}
	// The go.work file may not be "known" because we first access it through the
	// session. As a result, treat changes to the view's go.work file as always
	// relevant, even if they are only on-disk changes.
	//
	// TODO(rfindley): Make sure the go.work files are always known
	// to the view.
	if gowork, _ := v.GOWORK(); gowork == c.URI {
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

	// s.ignoreFilterOnce.Do(func() {
	// 	var dirs []string
	// 	for _, entry := range filepath.SplitList(s.view.gopath) {
	// 		dirs = append(dirs, filepath.Join(entry, "src"))
	// 	}

	// 	s.ignoreFilter = newIgnoreFilter(dirs)
	// })

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

	s.loadWorkspace(ctx, firstAttempt)
}

func (s *snapshot) loadWorkspace(ctx context.Context, firstAttempt bool) (loadErr error) {
	// A failure is retryable if it may have been due to context cancellation,
	// and this is not the initial workspace load (firstAttempt==true).
	//
	// The IWL runs on a detached context with a long (~10m) timeout, so
	// if the context was canceled we consider loading to have failed
	// permanently.
	retryableFailure := func() bool {
		return loadErr != nil && ctx.Err() != nil && !firstAttempt
	}
	defer func() {
		if !retryableFailure() {
			s.mu.Lock()
			s.initialized = true
			s.mu.Unlock()
		}
		if firstAttempt {
			close(s.view.initialWorkspaceLoad)
		}
	}()

	// TODO(rFindley): we should only locate template files on the first attempt,
	// or guard it via a different mechanism.
	s.locateTemplateFiles(ctx)

	// Collect module paths to load by parsing go.mod files. If a module fails to
	// parse, capture the parsing failure as a critical diagnostic.
	var scopes []loadScope                  // scopes to load
	var modDiagnostics []*source.Diagnostic // diagnostics for broken go.mod files
	addError := func(uri span.URI, err error) {
		modDiagnostics = append(modDiagnostics, &source.Diagnostic{
			URI:      uri,
			Severity: protocol.SeverityError,
			Source:   source.ListError,
			Message:  err.Error(),
		})
	}

	// TODO(rfindley): this should be predicated on the s.view.moduleMode().
	// There is no point loading ./... if we have an empty go.work.
	if len(s.workspaceModFiles) > 0 {
		for modURI := range s.workspaceModFiles {
			// Verify that the modfile is valid before trying to load it.
			//
			// TODO(rfindley): now that we no longer need to parse the modfile in
			// order to load scope, we could move these diagnostics to a more general
			// location where we diagnose problems with modfiles or the workspace.
			//
			// Be careful not to add context cancellation errors as critical module
			// errors.
			fh, err := s.ReadFile(ctx, modURI)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				addError(modURI, err)
				continue
			}
			parsed, err := s.ParseMod(ctx, fh)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				addError(modURI, err)
				continue
			}
			if parsed.File == nil || parsed.File.Module == nil {
				addError(modURI, fmt.Errorf("no module path for %s", modURI))
				continue
			}
			moduleDir := filepath.Dir(modURI.Filename())
			// Previously, we loaded <modulepath>/... for each module path, but that
			// is actually incorrect when the pattern may match packages in more than
			// one module. See golang/go#59458 for more details.
			scopes = append(scopes, moduleLoadScope{dir: moduleDir, modulePath: parsed.File.Module.Mod.Path})
		}
	} else {
		scopes = append(scopes, viewLoadScope("LOAD_VIEW"))
	}

	// If we're loading anything, ensure we also load builtin,
	// since it provides fake definitions (and documentation)
	// for types like int that are used everywhere.
	if len(scopes) > 0 {
		scopes = append(scopes, packageLoadScope("builtin"))
	}
	loadErr = s.load(ctx, true, scopes...)

	if retryableFailure() {
		return loadErr
	}

	var criticalErr *source.CriticalError
	switch {
	case loadErr != nil && ctx.Err() != nil:
		event.Error(ctx, fmt.Sprintf("initial workspace load: %v", loadErr), loadErr)
		criticalErr = &source.CriticalError{
			MainError: loadErr,
		}
	case loadErr != nil:
		event.Error(ctx, "initial workspace load failed", loadErr)
		extractedDiags := s.extractGoCommandErrors(ctx, loadErr)
		criticalErr = &source.CriticalError{
			MainError:   loadErr,
			Diagnostics: append(modDiagnostics, extractedDiags...),
		}
	case len(modDiagnostics) == 1:
		criticalErr = &source.CriticalError{
			MainError:   fmt.Errorf(modDiagnostics[0].Message),
			Diagnostics: modDiagnostics,
		}
	case len(modDiagnostics) > 1:
		criticalErr = &source.CriticalError{
			MainError:   fmt.Errorf("error loading module names"),
			Diagnostics: modDiagnostics,
		}
	}

	// Lock the snapshot when setting the initialized error.
	s.mu.Lock()
	defer s.mu.Unlock()
	s.initializedErr = criticalErr
	return loadErr
}

// invalidateContent invalidates the content of a Go file,
// including any position and type information that depends on it.
//
// invalidateContent returns a non-nil snapshot for the new content, along with
// a callback which the caller must invoke to release that snapshot.
//
// newOptions may be nil, in which case options remain unchanged.
func (v *View) invalidateContent(ctx context.Context, changes map[span.URI]source.FileHandle, newOptions *source.Options, forceReloadMetadata bool) (*snapshot, func()) {
	// Detach the context so that content invalidation cannot be canceled.
	ctx = xcontext.Detach(ctx)

	// This should be the only time we hold the view's snapshot lock for any period of time.
	v.snapshotMu.Lock()
	defer v.snapshotMu.Unlock()

	prevSnapshot, prevReleaseSnapshot := v.snapshot, v.releaseSnapshot

	if prevSnapshot == nil {
		panic("invalidateContent called after shutdown")
	}

	// Cancel all still-running previous requests, since they would be
	// operating on stale data.
	prevSnapshot.cancel()

	// Do not clone a snapshot until its view has finished initializing.
	prevSnapshot.AwaitInitialized(ctx)

	// Save one lease of the cloned snapshot in the view.
	v.snapshot, v.releaseSnapshot = prevSnapshot.clone(ctx, v.baseCtx, changes, newOptions, forceReloadMetadata)

	prevReleaseSnapshot()
	v.destroy(prevSnapshot, "View.invalidateContent")

	// Return a second lease to the caller.
	return v.snapshot, v.snapshot.Acquire()
}

func (s *Session) getWorkspaceInformation(ctx context.Context, folder span.URI, options *source.Options) (workspaceInformation, error) {
	if err := checkPathCase(folder.Filename()); err != nil {
		return workspaceInformation{}, fmt.Errorf("invalid workspace folder path: %w; check that the casing of the configured workspace folder path agrees with the casing reported by the operating system", err)
	}
	var err error
	info := workspaceInformation{
		folder: folder,
	}
	inv := gocommand.Invocation{
		WorkingDir: folder.Filename(),
		Env:        options.EnvSlice(),
	}
	info.goversion, err = gocommand.GoVersion(ctx, inv, s.gocmdRunner)
	if err != nil {
		return info, err
	}
	info.goversionOutput, err = gocommand.GoVersionOutput(ctx, inv, s.gocmdRunner)
	if err != nil {
		return info, err
	}
	if err := info.load(ctx, folder.Filename(), options.EnvSlice(), s.gocmdRunner); err != nil {
		return info, err
	}
	// The value of GOPACKAGESDRIVER is not returned through the go command.
	gopackagesdriver := os.Getenv("GOPACKAGESDRIVER")
	// A user may also have a gopackagesdriver binary on their machine, which
	// works the same way as setting GOPACKAGESDRIVER.
	tool, _ := exec.LookPath("gopackagesdriver")
	info.hasGopackagesDriver = gopackagesdriver != "off" && (gopackagesdriver != "" || tool != "")

	// filterFunc is the path filter function for this workspace folder. Notably,
	// it is relative to folder (which is specified by the user), not root.
	filterFunc := pathExcludedByFilterFunc(folder.Filename(), info.gomodcache, options)
	info.gomod, err = findWorkspaceModFile(ctx, folder, s, filterFunc)
	if err != nil {
		return info, err
	}

	// Check if the workspace is within any GOPATH directory.
	for _, gp := range filepath.SplitList(info.gopath) {
		if source.InDir(filepath.Join(gp, "src"), folder.Filename()) {
			info.inGOPATH = true
			break
		}
	}

	// Compute the "working directory", which is where we run go commands.
	//
	// Note: if gowork is in use, this will default to the workspace folder. In
	// the past, we would instead use the folder containing go.work. This should
	// not make a difference, and in fact may improve go list error messages.
	//
	// TODO(golang/go#57514): eliminate the expandWorkspaceToModule setting
	// entirely.
	if options.ExpandWorkspaceToModule && info.gomod != "" {
		info.goCommandDir = span.URIFromPath(filepath.Dir(info.gomod.Filename()))
	} else {
		info.goCommandDir = folder
	}
	return info, nil
}

// findWorkspaceModFile searches for a single go.mod file relative to the given
// folder URI, using the following algorithm:
//  1. if there is a go.mod file in a parent directory, return it
//  2. else, if there is exactly one nested module, return it
//  3. else, return ""
func findWorkspaceModFile(ctx context.Context, folderURI span.URI, fs source.FileSource, excludePath func(string) bool) (span.URI, error) {
	folder := folderURI.Filename()
	match, err := findRootPattern(ctx, folder, "go.mod", fs)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return "", ctxErr
		}
		return "", err
	}
	if match != "" {
		return span.URIFromPath(match), nil
	}

	// ...else we should check if there's exactly one nested module.
	all, err := findModules(folderURI, excludePath, 2)
	if err == errExhausted {
		// Fall-back behavior: if we don't find any modules after searching 10000
		// files, assume there are none.
		event.Log(ctx, fmt.Sprintf("stopped searching for modules after %d files", fileLimit))
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if len(all) == 1 {
		// range to access first element.
		for uri := range all {
			return uri, nil
		}
	}
	return "", nil
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

func pathExcludedByFilterFunc(folder, gomodcache string, opts *source.Options) func(string) bool {
	filterer := buildFilterer(folder, gomodcache, opts)
	return func(path string) bool {
		return pathExcludedByFilter(path, filterer)
	}
}

// pathExcludedByFilter reports whether the path (relative to the workspace
// folder) should be excluded by the configured directory filters.
//
// TODO(rfindley): passing root and gomodcache here makes it confusing whether
// path should be absolute or relative, and has already caused at least one
// bug.
func pathExcludedByFilter(path string, filterer *source.Filterer) bool {
	path = strings.TrimPrefix(filepath.ToSlash(path), "/")
	return filterer.Disallow(path)
}

func buildFilterer(folder, gomodcache string, opts *source.Options) *source.Filterer {
	filters := opts.DirectoryFilters

	if pref := strings.TrimPrefix(gomodcache, folder); pref != gomodcache {
		modcacheFilter := "-" + strings.TrimPrefix(filepath.ToSlash(pref), "/")
		filters = append(filters, modcacheFilter)
	}
	return source.NewFilterer(filters)
}
