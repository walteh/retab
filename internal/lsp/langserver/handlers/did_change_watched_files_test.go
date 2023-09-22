// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

// func TestLangServer_DidChangeWatchedFiles_change_file(t *testing.T) {
// 	tmpDir := TempDir(t)

//
// 	originalSrc := `variable "original" {
//   default = "foo"
// }
// `
// 	err := os.WriteFile(filepath.Join(tmpDir.Path(), "main.tf"), []byte(originalSrc), 0o755)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	ss, err := state.NewStateStore()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
// 		StateStore:      ss,
//
// 	}))
// 	stop := ls.Start(t)
// 	defer stop()

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "initialize",
// 		ReqParams: fmt.Sprintf(`{
// 	    "capabilities": {},
// 	    "rootUri": %q,
// 	    "processId": 12345
// 	}`, tmpDir.URI)})
//
// 	ls.Notify(t, &langserver.CallRequest{
// 		Method:    "initialized",
// 		ReqParams: "{}",
// 	})

// 	// Verify main.tf was parsed
// 	mod, err := ss.DocumentStore.GetDocument(document.HandleFromPath(tmpDir.Path()))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles := mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok := parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(originalSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}

// 	// Change main.tf on disk
// 	newSrc := `variable "new" {
//   default = "foo"
// }
// `
// 	err = os.WriteFile(filepath.Join(tmpDir.Path(), "main.tf"), []byte(newSrc), 0o755)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Verify nothing has changed yet
// 	mod, err = ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles = mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok = parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(originalSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "workspace/didChangeWatchedFiles",
// 		ReqParams: fmt.Sprintf(`{
//     "changes": [
//         {
//             "uri": "%s/main.tf",
//             "type": 2
//         }
//     ]
// }`, TempDir(t).URI)})

// 	// Verify file was re-parsed
// 	mod, err = ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles = mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok = parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(newSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}
// }

// func TestLangServer_DidChangeWatchedFiles_create_file(t *testing.T) {
// 	tmpDir := TempDir(t)

//
// 	originalSrc := `variable "original" {
//   default = "foo"
// }
// `
// 	err := os.WriteFile(filepath.Join(tmpDir.Path(), "main.tf"), []byte(originalSrc), 0o755)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	ss, err := state.NewStateStore()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
// 		StateStore:      ss,
//
// 	}))
// 	stop := ls.Start(t)
// 	defer stop()

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "initialize",
// 		ReqParams: fmt.Sprintf(`{
// 	    "capabilities": {},
// 	    "rootUri": %q,
// 	    "processId": 12345
// 	}`, tmpDir.URI)})
//
// 	ls.Notify(t, &langserver.CallRequest{
// 		Method:    "initialized",
// 		ReqParams: "{}",
// 	})

// 	// Verify main.tf was parsed
// 	mod, err := ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles := mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok := parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(originalSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}

// 	// Create another.tf on disk
// 	newSrc := `variable "another" {
//   default = "foo"
// }
// `
// 	err = os.WriteFile(filepath.Join(tmpDir.Path(), "another.tf"), []byte(newSrc), 0o755)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Verify another.tf was not parsed *yet*
// 	mod, err = ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles = mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok = parsedFiles["another.tf"]
// 	if ok {
// 		t.Fatalf("not expected to be parsed: %q", "another.tf")
// 	}

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "workspace/didChangeWatchedFiles",
// 		ReqParams: fmt.Sprintf(`{
//     "changes": [
//         {
//             "uri": "%s/main.tf",
//             "type": 1
//         }
//     ]
// }`, TempDir(t).URI)})
//

// 	// Verify another.tf was parsed
// 	mod, err = ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles = mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok = parsedFiles["another.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "another.tf")
// 	}
// 	if diff := cmp.Diff(newSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "another.tf", diff)
// 	}
// }

// func TestLangServer_DidChangeWatchedFiles_delete_file(t *testing.T) {
// 	tmpDir := TempDir(t)

//
// 	originalSrc := `variable "original" {
//   default = "foo"
// }
// `
// 	err := os.WriteFile(filepath.Join(tmpDir.Path(), "main.tf"), []byte(originalSrc), 0o755)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	ss, err := state.NewStateStore()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
// 		StateStore:      ss,
//
// 	}))
// 	stop := ls.Start(t)
// 	defer stop()

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "initialize",
// 		ReqParams: fmt.Sprintf(`{
// 	    "capabilities": {},
// 	    "rootUri": %q,
// 	    "processId": 12345
// 	}`, tmpDir.URI)})
//
// 	ls.Notify(t, &langserver.CallRequest{
// 		Method:    "initialized",
// 		ReqParams: "{}",
// 	})

// 	// Verify main.tf was parsed
// 	mod, err := ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles := mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok := parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(originalSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}

// 	// Delete main.tf from disk
// 	err = os.Remove(filepath.Join(tmpDir.Path(), "main.tf"))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Verify main.tf still remains parsed
// 	mod, err = ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles = mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok = parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(originalSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "workspace/didChangeWatchedFiles",
// 		ReqParams: fmt.Sprintf(`{
//     "changes": [
//         {
//             "uri": "%s/main.tf",
//             "type": 3
//         }
//     ]
// }`, TempDir(t).URI)})

// 	// Verify main.tf was deleted
// 	mod, err = ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles = mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok = parsedFiles["main.tf"]
// 	if ok {
// 		t.Fatalf("not expected file to be parsed: %q", "main.tf")
// 	}
// }

// func TestLangServer_DidChangeWatchedFiles_change_dir(t *testing.T) {
// 	tmpDir := TempDir(t)

//
// 	originalSrc := `variable "original" {
//   default = "foo"
// }
// `
// 	err := os.WriteFile(filepath.Join(tmpDir.Path(), "main.tf"), []byte(originalSrc), 0o755)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	ss, err := state.NewStateStore()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
// 		StateStore:      ss,
//
// 	}))
// 	stop := ls.Start(t)
// 	defer stop()

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "initialize",
// 		ReqParams: fmt.Sprintf(`{
// 	    "capabilities": {},
// 	    "rootUri": %q,
// 	    "processId": 12345
// 	}`, tmpDir.URI)})
//
// 	ls.Notify(t, &langserver.CallRequest{
// 		Method:    "initialized",
// 		ReqParams: "{}",
// 	})

// 	// Verify main.tf was parsed
// 	mod, err := ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles := mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok := parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(originalSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}

// 	// Change main.tf on disk
// 	newSrc := `variable "new" {
//   default = "foo"
// }
// `
// 	err = os.WriteFile(filepath.Join(tmpDir.Path(), "main.tf"), []byte(newSrc), 0o755)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Verify nothing has changed yet
// 	mod, err = ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles = mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok = parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(originalSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "workspace/didChangeWatchedFiles",
// 		ReqParams: fmt.Sprintf(`{
//     "changes": [
//         {
//             "uri": %q,
//             "type": 2
//         }
//     ]
// }`, TempDir(t).URI)})

// 	// Verify file was re-parsed
// 	mod, err = ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles = mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok = parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(newSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}
// }

// func TestLangServer_DidChangeWatchedFiles_create_dir(t *testing.T) {
// 	tmpDir := TempDir(t)

//
// 	originalSrc := `variable "original" {
//   default = "foo"
// }
// `
// 	err := os.WriteFile(filepath.Join(tmpDir.Path(), "main.tf"), []byte(originalSrc), 0o755)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	ss, err := state.NewStateStore()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
// 		StateStore:      ss,
//
// 	}))
// 	stop := ls.Start(t)
// 	defer stop()

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "initialize",
// 		ReqParams: fmt.Sprintf(`{
// 	    "capabilities": {},
// 	    "rootUri": %q,
// 	    "processId": 12345
// 	}`, tmpDir.URI)})
//
// 	ls.Notify(t, &langserver.CallRequest{
// 		Method:    "initialized",
// 		ReqParams: "{}",
// 	})

// 	// Verify main.tf was parsed
// 	mod, err := ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles := mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok := parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(originalSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}

// 	// Create new ./submodule w/ main.tf on disk
// 	submodPath := filepath.Join(tmpDir.Path(), "submodule")
// 	submodHandle := document.DirHandleFromPath(submodPath)
// 	err = os.Mkdir(submodPath, 0o755)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	newSrc := `variable "new" {
//   default = "foo"
// }
// `
// 	err = os.WriteFile(filepath.Join(submodPath, "main.tf"), []byte(newSrc), 0o755)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	// Verify submodule was not parsed yet
// 	mod, err = ss.Modules.ModuleByPath(submodPath)
// 	if err == nil {
// 		t.Fatalf("%q: expected module not to be found", submodPath)
// 	}

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "workspace/didChangeWatchedFiles",
// 		ReqParams: fmt.Sprintf(`{
//     "changes": [
//         {
//             "uri": %q,
//             "type": 1
//         }
//     ]
// }`, submodHandle.URI)})
//

// 	// Verify submodule was parsed
// 	mod, err = ss.Modules.ModuleByPath(submodPath)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles = mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok = parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(newSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}
// }

// func TestLangServer_DidChangeWatchedFiles_delete_dir(t *testing.T) {
// 	tmpDir := TempDir(t)

//
// 	originalSrc := `variable "original" {
//   default = "foo"
// }
// `
// 	err := os.WriteFile(filepath.Join(tmpDir.Path(), "main.tf"), []byte(originalSrc), 0o755)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	ss, err := state.NewStateStore()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
// 		StateStore:      ss,
//
// 	}))
// 	stop := ls.Start(t)
// 	defer stop()

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "initialize",
// 		ReqParams: fmt.Sprintf(`{
// 	    "capabilities": {},
// 	    "rootUri": %q,
// 	    "processId": 12345
// 	}`, tmpDir.URI)})
//
// 	ls.Notify(t, &langserver.CallRequest{
// 		Method:    "initialized",
// 		ReqParams: "{}",
// 	})

// 	// Verify main.tf was parsed
// 	mod, err := ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles := mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok := parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(originalSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}

// 	// Delete directory from disk
// 	err = os.RemoveAll(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Verify nothing has changed yet
// 	mod, err = ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	parsedFiles = mod.ParsedModuleFiles.AsMap()
// 	parsedFile, ok = parsedFiles["main.tf"]
// 	if !ok {
// 		t.Fatalf("file not parsed: %q", "main.tf")
// 	}
// 	if diff := cmp.Diff(originalSrc, string(parsedFile.Bytes)); diff != "" {
// 		t.Fatalf("bytes mismatch for %q: %s", "main.tf", diff)
// 	}

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "workspace/didChangeWatchedFiles",
// 		ReqParams: fmt.Sprintf(`{
//     "changes": [
//         {
//             "uri": %q,
//             "type": 3
//         }
//     ]
// }`, TempDir(t).URI)})

// 	// Verify module is gone
// 	_, err = ss.Modules.ModuleByPath(tmpDir.Path())
// 	if err == nil {
// 		t.Fatalf("expected module at %q to be gone", tmpDir.Path())
// 	}
// }

// func TestLangServer_DidChangeWatchedFiles_pluginChange(t *testing.T) {
// 	testData, err := filepath.Abs("testdata")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	originalTestDir := filepath.Join(testData, "single-fake-provider")
// 	testDir := t.TempDir()
// 	// Copy test configuration so the test can run in isolation
// 	err = copy.Copy(originalTestDir, testDir)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	testHandle := document.DirHandleFromPath(testDir)

// 	ss, err := state.NewStateStore()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
// 		StateStore:      ss,
//
// 	}))
// 	stop := ls.Start(t)
// 	defer stop()

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "initialize",
// 		ReqParams: fmt.Sprintf(`{
// 	    "capabilities": {},
// 	    "rootUri": %q,
// 	    "processId": 12345
// 	}`, testHandle.URI)})
//
// 	ls.Notify(t, &langserver.CallRequest{
// 		Method:    "initialized",
// 		ReqParams: "{}",
// 	})

// 	addr := tfaddr.MustParseProviderSource("-/foo")
// 	vc := version.MustConstraints(version.NewConstraint(">= 1.0"))

// 	_, err = ss.ProviderSchemas.ProviderSchema(testHandle.Path(), addr, vc)
// 	if err == nil {
// 		t.Fatal("expected -/foo schema to be missing")
// 	}

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "workspace/didChangeWatchedFiles",
// 		ReqParams: fmt.Sprintf(`{
//     "changes": [
//         {
//             "uri": "%s/.terraform.lock.hcl",
//             "type": 1
//         }
//     ]
// }`, testHandle.URI)})

// 	_, err = ss.ProviderSchemas.ProviderSchema(testHandle.Path(), addr, vc)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }

// func TestLangServer_DidChangeWatchedFiles_moduleInstalled(t *testing.T) {
// 	testData, err := filepath.Abs("testdata")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	originalTestDir := filepath.Join(testData, "uninitialized-single-submodule")
// 	testDir := t.TempDir()
// 	// Copy test configuration so the test can run in isolation
// 	err = copy.Copy(originalTestDir, testDir)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	testHandle := document.DirHandleFromPath(testDir)

// 	ss, err := state.NewStateStore()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
// 		StateStore:      ss,
//
// 	}))
// 	stop := ls.Start(t)
// 	defer stop()

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "initialize",
// 		ReqParams: fmt.Sprintf(`{
// 	    "capabilities": {},
// 	    "rootUri": %q,
// 	    "processId": 12345
// 	}`, testHandle.URI)})
//
// 	ls.Notify(t, &langserver.CallRequest{
// 		Method:    "initialized",
// 		ReqParams: "{}",
// 	})

// 	submodulePath := filepath.Join(testDir, ".terraform", "modules", "azure-hcp-consul")
// 	_, err = ss.Modules.ModuleByPath(submodulePath)
// 	if err == nil || !state.IsModuleNotFound(err) {
// 		t.Fatalf("expected submodule not to be found: %s", err)
// 	}

// 	// Install Terraform
// 	tfVersion := version.Must(version.NewVersion("1.1.7"))
// 	i := install.NewInstaller()
// 	ctx := context.Background()
// 	execPath, err := i.Install(ctx, []src.Installable{
// 		&releases.ExactVersion{
// 			Product: product.Terraform,
// 			Version: tfVersion,
// 		},
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Install submodule
// 	tf, err := exec.NewExecutor(testHandle.Path(), execPath)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = tf.Get(ctx)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	ls.Call(t, &langserver.CallRequest{
// 		Method: "workspace/didChangeWatchedFiles",
// 		ReqParams: fmt.Sprintf(`{
//     "changes": [
//         {
//             "uri": "%s/.terraform/modules/modules.json",
//             "type": 1
//         }
//     ]
// }`, testHandle.URI)})

// 	mod, err := ss.Modules.ModuleByPath(submodulePath)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if len(mod.Meta.Variables) != 8 {
// 		t.Fatalf("expected exactly 8 variables, %d given", len(mod.Meta.Variables))
// 	}
// }
