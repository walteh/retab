// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cache

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/walteh/retab/gen/gopls/bug"
	"github.com/walteh/retab/gen/gopls/event"
	"github.com/walteh/retab/gen/gopls/memoize"
	"github.com/walteh/retab/gen/gopls/persistent"
	"github.com/walteh/retab/gen/gopls/span"
	"github.com/walteh/retab/internal/source"
)

var _ source.Snapshot = (*snapshot)(nil)

type snapshot struct {
	sequenceID uint64
	globalID   source.GlobalSnapshotID

	// TODO(rfindley): the snapshot holding a reference to the view poses
	// lifecycle problems: a view may be shut down and waiting for work
	// associated with this snapshot to complete. While most accesses of the view
	// are benign (options or workspace information), this is not formalized and
	// it is wrong for the snapshot to use a shutdown view.
	//
	// Fix this by passing options and workspace information to the snapshot,
	// both of which should be immutable for the snapshot.
	view *View

	cancel        func()
	backgroundCtx context.Context

	store *memoize.Store // cache of handles shared by all snapshots

	refcount    sync.WaitGroup // number of references
	destroyedBy *string        // atomically set to non-nil in Destroy once refcount = 0

	// initialized reports whether the snapshot has been initialized. Concurrent
	// initialization is guarded by the view.initializationSema. Each snapshot is
	// initialized at most once: concurrent initialization is guarded by
	// view.initializationSema.
	initialized bool
	// initializedErr holds the last error resulting from initialization. If
	// initialization fails, we only retry when the workspace modules change,
	// to avoid too many go/packages calls.
	initializedErr *source.CriticalError

	// mu guards all of the maps in the snapshot, as well as the builtin URI.
	mu sync.Mutex

	// builtin is the location of builtin.go in GOROOT.
	//
	// TODO(rfindley): would it make more sense to eagerly parse builtin, and
	// instead store a *ParsedGoFile here?
	builtin span.URI

	// // meta holds loaded metadata.
	// //
	// // meta is guarded by mu, but the metadataGraph itself is immutable.
	// // TODO(rfindley): in many places we hold mu while operating on meta, even
	// // though we only need to hold mu while reading the pointer.
	// meta *metadataGraph

	// files maps file URIs to their corresponding FileHandles.
	// It may invalidated when a file's content changes.
	files *fileMap

	// symbolizeHandles maps each file URI to a handle for the future
	// result of computing the symbols declared in that file.
	symbolizeHandles *persistent.Map[span.URI, *memoize.Promise] // *memoize.Promise[symbolizeResult]

	// // packages maps a packageKey to a *packageHandle.
	// // It may be invalidated when a file's content changes.
	// //
	// // Invariants to preserve:
	// //  - packages.Get(id).meta == meta.metadata[id] for all ids
	// //  - if a package is in packages, then all of its dependencies should also
	// //    be in packages, unless there is a missing import
	// packages *persistent.Map[PackageID, *packageHandle]

	// // activePackages maps a package ID to a memoized active package, or nil if
	// // the package is known not to be open.
	// //
	// // IDs not contained in the map are not known to be open or not open.
	// activePackages *persistent.Map[PackageID, *Package]

	// // workspacePackages contains the workspace's packages, which are loaded
	// // when the view is created. It contains no intermediate test variants.
	// // TODO(rfindley): use a persistent.Map.
	// workspacePackages map[PackageID]PackagePath

	// // shouldLoad tracks packages that need to be reloaded, mapping a PackageID
	// // to the package paths that should be used to reload it
	// //
	// // When we try to load a package, we clear it from the shouldLoad map
	// // regardless of whether the load succeeded, to prevent endless loads.
	// shouldLoad map[PackageID][]PackagePath

	// unloadableFiles keeps track of files that we've failed to load.
	unloadableFiles *persistent.Set[span.URI]

	// Only compute module prefixes once, as they are used with high frequency to
	// detect ignored files.
	ignoreFilterOnce sync.Once
	ignoreFilter     *ignoreFilter

	// typeCheckMu guards type checking.
	//
	// Only one type checking pass should be running at a given time, for two reasons:
	//  1. type checking batches are optimized to use all available processors.
	//     Generally speaking, running two type checking batches serially is about
	//     as fast as running them in parallel.
	//  2. type checking produces cached artifacts that may be re-used by the
	//     next type-checking batch: the shared import graph and the set of
	//     active packages. Running type checking batches in parallel after an
	//     invalidation can cause redundant calculation of this shared state.
	typeCheckMu sync.Mutex

	// options holds the user configuration at the time this snapshot was
	// created.
	options *source.Options
}

var globalSnapshotID uint64

func nextSnapshotID() source.GlobalSnapshotID {
	return source.GlobalSnapshotID(atomic.AddUint64(&globalSnapshotID, 1))
}

var _ memoize.RefCounted = (*snapshot)(nil) // snapshots are reference-counted

// Acquire prevents the snapshot from being destroyed until the returned function is called.
//
// (s.Acquire().release() could instead be expressed as a pair of
// method calls s.IncRef(); s.DecRef(). The latter has the advantage
// that the DecRefs are fungible and don't require holding anything in
// addition to the refcounted object s, but paradoxically that is also
// an advantage of the current approach, which forces the caller to
// consider the release function at every stage, making a reference
// leak more obvious.)
func (s *snapshot) Acquire() func() {
	type uP = unsafe.Pointer
	if destroyedBy := atomic.LoadPointer((*uP)(uP(&s.destroyedBy))); destroyedBy != nil {
		log.Panicf("%d: acquire() after Destroy(%q)", s.globalID, *(*string)(destroyedBy))
	}
	s.refcount.Add(1)
	return s.refcount.Done
}

func (s *snapshot) awaitPromise(ctx context.Context, p *memoize.Promise) (interface{}, error) {
	return p.Get(ctx, s)
}

// destroy waits for all leases on the snapshot to expire then releases
// any resources (reference counts and files) associated with it.
// Snapshots being destroyed can be awaited using v.destroyWG.
//
// TODO(adonovan): move this logic into the release function returned
// by Acquire when the reference count becomes zero. (This would cost
// us the destroyedBy debug info, unless we add it to the signature of
// memoize.RefCounted.Acquire.)
//
// The destroyedBy argument is used for debugging.
//
// v.snapshotMu must be held while calling this function, in order to preserve
// the invariants described by the docstring for v.snapshot.
func (v *View) destroy(s *snapshot, destroyedBy string) {
	v.snapshotWG.Add(1)
	go func() {
		defer v.snapshotWG.Done()
		s.destroy(destroyedBy)
	}()
}

func (s *snapshot) destroy(destroyedBy string) {
	// Wait for all leases to end before commencing destruction.
	s.refcount.Wait()

	// Report bad state as a debugging aid.
	// Not foolproof: another thread could acquire() at this moment.
	type uP = unsafe.Pointer // looking forward to generics...
	if old := atomic.SwapPointer((*uP)(uP(&s.destroyedBy)), uP(&destroyedBy)); old != nil {
		log.Panicf("%d: Destroy(%q) after Destroy(%q)", s.globalID, destroyedBy, *(*string)(old))
	}

	// s.packages.Destroy()
	// s.activePackages.Destroy()
	s.files.Destroy()
	s.symbolizeHandles.Destroy()

	// s.modTidyHandles.Destroy()
	// s.modVulnHandles.Destroy()
	// s.modWhyHandles.Destroy()
	s.unloadableFiles.Destroy()
}

func (s *snapshot) SequenceID() uint64 {
	return s.sequenceID
}

func (s *snapshot) GlobalID() source.GlobalSnapshotID {
	return s.globalID
}

func (s *snapshot) View() source.View {
	return s.view
}

func (s *snapshot) FileKind(fh source.FileHandle) source.FileKind {
	// The kind of an unsaved buffer comes from the
	// TextDocumentItem.LanguageID field in the didChange event,
	// not from the file name. They may differ.
	if o, ok := fh.(*Overlay); ok {
		if o.kind != source.UnknownKind {
			return o.kind
		}
	}

	fext := filepath.Ext(fh.URI().Filename())
	switch fext {
	case ".go":
		return source.Go
	case ".mod":
		return source.Mod
	case ".sum":
		return source.Sum
	case ".work":
		return source.Work
	}
	exts := s.options.TemplateExtensions
	for _, ext := range exts {
		if fext == ext || fext == "."+ext {
			return source.Tmpl
		}
	}
	// and now what? This should never happen, but it does for cgo before go1.15
	return source.Go
}

func (s *snapshot) Options() *source.Options {
	return s.options // temporarily return view options.
}

func (s *snapshot) BackgroundContext() context.Context {
	return s.backgroundCtx
}

func (s *snapshot) Templates() map[span.URI]source.FileHandle {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmpls := map[span.URI]source.FileHandle{}
	s.files.Range(func(k span.URI, fh source.FileHandle) {
		if s.FileKind(fh) == source.Tmpl {
			tmpls[k] = fh
		}
	})
	return tmpls
}

// workspaceMode describes the way in which the snapshot's workspace should
// be loaded.
//
// TODO(rfindley): remove this, in favor of specific methods.
func (s *snapshot) workspaceMode() workspaceMode {
	var mode workspaceMode

	return mode
}

func (s *snapshot) buildOverlay() map[string][]byte {
	overlays := make(map[string][]byte)
	for _, overlay := range s.overlays() {
		if overlay.saved {
			continue
		}
		// TODO(rfindley): previously, there was a todo here to make sure we don't
		// send overlays outside of the current view. IMO we should instead make
		// sure this doesn't matter.
		overlays[overlay.URI().Filename()] = overlay.content
	}
	return overlays
}

func (s *snapshot) overlays() []*Overlay {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.files.Overlays()
}

func boolLess(x, y bool) bool { return !x && y } // false < true

// watchSubdirs reports whether gopls should request separate file watchers for
// each relevant subdirectory. This is necessary only for clients (namely VS
// Code) that do not send notifications for individual files in a directory
// when the entire directory is deleted.
func (s *snapshot) watchSubdirs() bool {
	switch p := s.options.SubdirWatchPatterns; p {
	case source.SubdirWatchPatternsOn:
		return true
	case source.SubdirWatchPatternsOff:
		return false
	case source.SubdirWatchPatternsAuto:
		// See the documentation of InternalOptions.SubdirWatchPatterns for an
		// explanation of why VS Code gets a different default value here.
		//
		// Unfortunately, there is no authoritative list of client names, nor any
		// requirements that client names do not change. We should update the VS
		// Code extension to set a default value of "subdirWatchPatterns" to "on",
		// so that this workaround is only temporary.
		if s.options.ClientInfo != nil && s.options.ClientInfo.Name == "Visual Studio Code" {
			return true
		}
		return false
	default:
		bug.Reportf("invalid subdirWatchPatterns: %q", p)
		return false
	}
}

// filesInDir returns all files observed by the snapshot that are contained in
// a directory with the provided URI.
func (s *snapshot) filesInDir(uri span.URI) []span.URI {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := uri.Filename()
	if !s.files.Dirs().Contains(dir) {
		return nil
	}
	var files []span.URI
	s.files.Range(func(uri span.URI, _ source.FileHandle) {
		if source.InDir(dir, uri.Filename()) {
			files = append(files, uri)
		}
	})
	return files
}

func moduleForURI(modFiles map[span.URI]struct{}, uri span.URI) span.URI {
	var match span.URI
	for modURI := range modFiles {
		if !source.InDir(filepath.Dir(modURI.Filename()), uri.Filename()) {
			continue
		}
		if len(modURI) > len(match) {
			match = modURI
		}
	}
	return match
}

func (s *snapshot) FindFile(uri span.URI) source.FileHandle {
	s.view.markKnown(uri)

	s.mu.Lock()
	defer s.mu.Unlock()

	result, _ := s.files.Get(uri)
	return result
}

// ReadFile returns a File for the given URI. If the file is unknown it is added
// to the managed set.
//
// ReadFile succeeds even if the file does not exist. A non-nil error return
// indicates some type of internal error, for example if ctx is cancelled.
func (s *snapshot) ReadFile(ctx context.Context, uri span.URI) (source.FileHandle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return lockedSnapshot{s}.ReadFile(ctx, uri)
}

// preloadFiles delegates to the view FileSource to read the requested uris in
// parallel, without holding the snapshot lock.
func (s *snapshot) preloadFiles(ctx context.Context, uris []span.URI) {
	files := make([]source.FileHandle, len(uris))
	var wg sync.WaitGroup
	iolimit := make(chan struct{}, 20) // I/O concurrency limiting semaphore
	for i, uri := range uris {
		wg.Add(1)
		iolimit <- struct{}{}
		go func(i int, uri span.URI) {
			defer wg.Done()
			fh, err := s.view.fs.ReadFile(ctx, uri)
			<-iolimit
			if err != nil && ctx.Err() == nil {
				event.Error(ctx, fmt.Sprintf("reading %s", uri), err)
				return
			}
			files[i] = fh
		}(i, uri)
	}
	wg.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, fh := range files {
		if fh == nil {
			continue // error logged above
		}
		uri := uris[i]
		if _, ok := s.files.Get(uri); !ok {
			s.files.Set(uri, fh)
		}
	}
}

// A lockedSnapshot implements the source.FileSource interface while holding
// the lock for the wrapped snapshot.
type lockedSnapshot struct{ wrapped *snapshot }

func (s lockedSnapshot) ReadFile(ctx context.Context, uri span.URI) (source.FileHandle, error) {
	s.wrapped.view.markKnown(uri)
	if fh, ok := s.wrapped.files.Get(uri); ok {
		return fh, nil
	}

	fh, err := s.wrapped.view.fs.ReadFile(ctx, uri)
	if err != nil {
		return nil, err
	}
	s.wrapped.files.Set(uri, fh)
	return fh, nil
}

func (s *snapshot) IsOpen(uri span.URI) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	fh, _ := s.files.Get(uri)
	_, open := fh.(*Overlay)
	return open
}

// TODO(rfindley): it would make sense for awaitLoaded to return metadata.
func (s *snapshot) awaitLoaded(ctx context.Context) error {
	loadErr := s.awaitLoadedAllErrors(ctx)

	// TODO(rfindley): eliminate this function as part of simplifying
	// CriticalErrors.
	if loadErr != nil {
		return loadErr.MainError
	}
	return nil
}

func (s *snapshot) awaitLoadedAllErrors(ctx context.Context) *source.CriticalError {
	// Do not return results until the snapshot's view has been initialized.
	s.AwaitInitialized(ctx)

	// TODO(rfindley): Should we be more careful about returning the
	// initialization error? Is it possible for the initialization error to be
	// corrected without a successful reinitialization?
	if err := s.getInitializationError(); err != nil {
		return err
	}

	// TODO(rfindley): revisit this handling. Calling reloadWorkspace with a
	// cancelled context should have the same effect, so this preemptive handling
	// should not be necessary.
	//
	// Also: GetCriticalError ignores context cancellation errors. Should we be
	// returning nil here?
	if ctx.Err() != nil {
		return &source.CriticalError{MainError: ctx.Err()}
	}

	// // TODO(rfindley): reloading is not idempotent: if we try to reload or load
	// // orphaned files below and fail, we won't try again. For that reason, we
	// // could get different results from subsequent calls to this function, which
	// // may cause critical errors to be suppressed.

	// if err := s.reloadWorkspace(ctx); err != nil {
	// 	diags := s.extractGoCommandErrors(ctx, err)
	// 	return &source.CriticalError{
	// 		MainError:   err,
	// 		Diagnostics: diags,
	// 	}
	// }

	// if err := s.reloadOrphanedOpenFiles(ctx); err != nil {
	// 	diags := s.extractGoCommandErrors(ctx, err)
	// 	return &source.CriticalError{
	// 		MainError:   err,
	// 		Diagnostics: diags,
	// 	}
	// }
	return nil
}

func (s *snapshot) getInitializationError() *source.CriticalError {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.initializedErr
}

func (s *snapshot) AwaitInitialized(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-s.view.initialWorkspaceLoad:
	}
	// We typically prefer to run something as intensive as the IWL without
	// blocking. I'm not sure if there is a way to do that here.
	s.initialize(ctx, false)
}

// TODO(golang/go#53756): this function needs to consider more than just the
// absolute URI, for example:
//   - the position of /vendor/ with respect to the relevant module root
//   - whether or not go.work is in use (as vendoring isn't supported in workspace mode)
//
// Most likely, each call site of inVendor needs to be reconsidered to
// understand and correctly implement the desired behavior.
func inVendor(uri span.URI) bool {
	_, after, found := strings.Cut(string(uri), "/vendor/")
	// Only subdirectories of /vendor/ are considered vendored
	// (/vendor/a/foo.go is vendored, /vendor/foo.go is not).
	return found && strings.Contains(after, "/")
}

func cloneWithout[V any](m *persistent.Map[span.URI, V], changes map[span.URI]source.FileHandle) *persistent.Map[span.URI, V] {
	m2 := m.Clone()
	for k := range changes {
		m2.Delete(k)
	}
	return m2
}

// deleteMostRelevantModFile deletes the mod file most likely to be the mod
// file for the changed URI, if it exists.
//
// Specifically, this is the longest mod file path in a directory containing
// changed. This might not be accurate if there is another mod file closer to
// changed that happens not to be present in the map, but that's OK: the goal
// of this function is to guarantee that IF the nearest mod file is present in
// the map, it is invalidated.
func deleteMostRelevantModFile(m *persistent.Map[span.URI, *memoize.Promise], changed span.URI) {
	var mostRelevant span.URI
	changedFile := changed.Filename()

	m.Range(func(modURI span.URI, _ *memoize.Promise) {
		if len(modURI) > len(mostRelevant) {
			if source.InDir(filepath.Dir(modURI.Filename()), changedFile) {
				mostRelevant = modURI
			}
		}
	})
	if mostRelevant != "" {
		m.Delete(mostRelevant)
	}
}

// fileWasSaved reports whether the FileHandle passed in has been saved. It
// accomplishes this by checking to see if the original and current FileHandles
// are both overlays, and if the current FileHandle is saved while the original
// FileHandle was not saved.
func fileWasSaved(originalFH, currentFH source.FileHandle) bool {
	c, ok := currentFH.(*Overlay)
	if !ok || c == nil {
		return true
	}
	o, ok := originalFH.(*Overlay)
	if !ok || o == nil {
		return c.saved
	}
	return !o.saved && c.saved
}

// metadataChanges detects features of the change from oldFH->newFH that may
// affect package metadata.
//
// It uses lockedSnapshot to access cached parse information. lockedSnapshot
// must be locked.
//
// The result parameters have the following meaning:
//   - invalidate means that package metadata for packages containing the file
//     should be invalidated.
//   - pkgFileChanged means that the file->package associates for the file have
//     changed (possibly because the file is new, or because its package name has
//     changed).
//   - importDeleted means that an import has been deleted, or we can't
//     determine if an import was deleted due to errors.
func metadataChanges(ctx context.Context, lockedSnapshot *snapshot, oldFH, newFH source.FileHandle) (invalidate, pkgFileChanged, importDeleted bool) {
	if oe, ne := oldFH != nil && fileExists(oldFH), fileExists(newFH); !oe || !ne { // existential changes
		changed := oe != ne
		return changed, changed, !ne // we don't know if an import was deleted
	}

	// If the file hasn't changed, there's no need to reload.
	if oldFH.FileIdentity() == newFH.FileIdentity() {
		return false, false, false
	}

	fset := token.NewFileSet()
	// Parse headers to compare package names and imports.
	oldHeads, oldErr := lockedSnapshot.view.parseCache.parseFiles(ctx, fset, source.ParseHeader, false, oldFH)
	newHeads, newErr := lockedSnapshot.view.parseCache.parseFiles(ctx, fset, source.ParseHeader, false, newFH)

	if oldErr != nil || newErr != nil {
		errChanged := (oldErr == nil) != (newErr == nil)
		return errChanged, errChanged, (newErr != nil) // we don't know if an import was deleted
	}

	oldHead := oldHeads[0]
	newHead := newHeads[0]

	// `go list` fails completely if the file header cannot be parsed. If we go
	// from a non-parsing state to a parsing state, we should reload.
	if oldHead.ParseErr != nil && newHead.ParseErr == nil {
		return true, true, true // We don't know what changed, so fall back on full invalidation.
	}

	// If a package name has changed, the set of package imports may have changed
	// in ways we can't detect here. Assume an import has been deleted.
	if oldHead.File.Name.Name != newHead.File.Name.Name {
		return true, true, true
	}

	// Check whether package imports have changed. Only consider potentially
	// valid imports paths.
	oldImports := validImports(oldHead.File.Imports)
	newImports := validImports(newHead.File.Imports)

	for path := range newImports {
		if _, ok := oldImports[path]; ok {
			delete(oldImports, path)
		} else {
			invalidate = true // a new, potentially valid import was added
		}
	}

	if len(oldImports) > 0 {
		invalidate = true
		importDeleted = true
	}

	// If the change does not otherwise invalidate metadata, get the full ASTs in
	// order to check magic comments.
	//
	// Note: if this affects performance we can probably avoid parsing in the
	// common case by first scanning the source for potential comments.
	if !invalidate {
		origFulls, oldErr := lockedSnapshot.view.parseCache.parseFiles(ctx, fset, source.ParseFull, false, oldFH)
		newFulls, newErr := lockedSnapshot.view.parseCache.parseFiles(ctx, fset, source.ParseFull, false, newFH)
		if oldErr == nil && newErr == nil {
			invalidate = magicCommentsChanged(origFulls[0].File, newFulls[0].File)
		} else {
			// At this point, we shouldn't ever fail to produce a ParsedGoFile, as
			// we're already past header parsing.
			bug.Reportf("metadataChanges: unparseable file %v (old error: %v, new error: %v)", oldFH.URI(), oldErr, newErr)
		}
	}

	return invalidate, pkgFileChanged, importDeleted
}

func magicCommentsChanged(original *ast.File, current *ast.File) bool {
	oldComments := extractMagicComments(original)
	newComments := extractMagicComments(current)
	if len(oldComments) != len(newComments) {
		return true
	}
	for i := range oldComments {
		if oldComments[i] != newComments[i] {
			return true
		}
	}
	return false
}

// validImports extracts the set of valid import paths from imports.
func validImports(imports []*ast.ImportSpec) map[string]struct{} {
	m := make(map[string]struct{})
	for _, spec := range imports {
		if path := spec.Path.Value; validImportPath(path) {
			m[path] = struct{}{}
		}
	}
	return m
}

func validImportPath(path string) bool {
	path, err := strconv.Unquote(path)
	if err != nil {
		return false
	}
	if path == "" {
		return false
	}
	if path[len(path)-1] == '/' {
		return false
	}
	return true
}

var buildConstraintOrEmbedRe = regexp.MustCompile(`^//(go:embed|go:build|\s*\+build).*`)

// extractMagicComments finds magic comments that affect metadata in f.
func extractMagicComments(f *ast.File) []string {
	var results []string
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if buildConstraintOrEmbedRe.MatchString(c.Text) {
				results = append(results, c.Text)
			}
		}
	}
	return results
}

func (s *snapshot) BuiltinFile(ctx context.Context) (*source.ParsedGoFile, error) {
	s.AwaitInitialized(ctx)

	s.mu.Lock()
	builtin := s.builtin
	s.mu.Unlock()

	if builtin == "" {
		return nil, fmt.Errorf("no builtin package for view %s", s.view.name)
	}

	fh, err := s.ReadFile(ctx, builtin)
	if err != nil {
		return nil, err
	}
	// For the builtin file only, we need syntactic object resolution
	// (since we can't type check).
	mode := source.ParseFull &^ source.SkipObjectResolution
	pgfs, err := s.view.parseCache.parseFiles(ctx, token.NewFileSet(), mode, false, fh)
	if err != nil {
		return nil, err
	}
	return pgfs[0], nil
}

func (s *snapshot) IsBuiltin(uri span.URI) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	// We should always get the builtin URI in a canonical form, so use simple
	// string comparison here. span.CompareURI is too expensive.
	return uri == s.builtin
}

func (s *snapshot) setBuiltin(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.builtin = span.URIFromPath(path)
}
