// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package handlers

import (
	"fmt"
	"testing"

	"github.com/walteh/retab/internal/langserver"
	"github.com/walteh/retab/internal/langserver/session"
	"github.com/walteh/retab/internal/state"
)

func TestLangServer_codeActionWithoutInitialization(t *testing.T) {
	ls := langserver.NewLangServerMock(t, NewMockSession(nil))
	stop := ls.Start(t)
	defer stop()

	ls.CallAndExpectError(t, &langserver.CallRequest{
		Method: "textDocument/codeAction",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "terraform",
			"text": "provider \"github\" {}",
			"uri": "%s/main.tf"
		}
	}`, TempDir(t).URI)}, session.SessionNotInitialized.Err())
}

func TestLangServer_codeAction_basic(t *testing.T) {
	tmpDir := TempDir(t)

	ss, err := state.NewStateStore()
	if err != nil {
		t.Fatal(err)
	}

	ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
		StateStore: ss,
	}))
	stop := ls.Start(t)
	defer stop()

	ls.Call(t, &langserver.CallRequest{
		Method: "initialize",
		ReqParams: fmt.Sprintf(`{
	    "capabilities": {},
	    "rootUri": %q,
	    "processId": 12345
	}`, tmpDir.URI)})

	ls.Notify(t, &langserver.CallRequest{
		Method:    "initialized",
		ReqParams: "{}",
	})
	ls.Call(t, &langserver.CallRequest{
		Method: "textDocument/didOpen",
		ReqParams: fmt.Sprintf(`{
		"textDocument": {
			"version": 0,
			"languageId": "terraform",
			"text": "provider  \"test\"   {\n\n      }\n",
			"uri": "%s/main.tf"
		}
	}`, tmpDir.URI)})
	waitForAllJobs(t, ss)

	ls.CallAndExpectResponse(t, &langserver.CallRequest{
		Method: "textDocument/codeAction",
		ReqParams: fmt.Sprintf(`{
			"textDocument": { "uri": "%s/main.tf" },
			"range": {
				"start": { "line": 0, "character": 0 },
				"end": { "line": 1, "character": 0 }
			},
			"context": { "diagnostics": [], "only": ["source.formatAll.terraform"] }
		}`, tmpDir.URI)}, fmt.Sprintf(`{
			"jsonrpc": "2.0",
			"id": 3,
			"result": [
				{
					"title": "Format Document",
					"kind": "source.formatAll.terraform",
					"edit":{
						"changes":{
							"%s/main.tf": [
								{
									"range": {
										"start": {
											"line": 0,
											"character": 0
										},
										"end": {
											"line": 1,
											"character": 0
										}
									},
									"newText": "provider \"test\" {\n"
								},
								{
									"range": {
										"start": {
											"line": 2,
											"character": 0
										},
										"end": {
											"line": 3,
											"character": 0
										}
									},
									"newText": "}\n"
								}
							]
						}
					}
				}
			]
		}`, tmpDir.URI))
}

func TestLangServer_codeAction_no_code_action_requested(t *testing.T) {
	tmpDir := TempDir(t)

	tests := []struct {
		name    string
		request *langserver.CallRequest
		want    string
	}{
		{
			name: "no code action requested",
			request: &langserver.CallRequest{
				Method: "textDocument/codeAction",
				ReqParams: fmt.Sprintf(`{
						"textDocument": { "uri": "%s/main.tf" },
						"range": {
							"start": { "line": 0, "character": 0 },
							"end": { "line": 1, "character": 0 }
						},
						"context": { "diagnostics": [], "only": [""] }
					}`, tmpDir.URI)},
			want: `{
				"jsonrpc": "2.0",
				"id": 3,
				"result": null
			}`,
		},
		{
			name: "source.formatAll.terraform code action requested",
			request: &langserver.CallRequest{
				Method: "textDocument/codeAction",
				ReqParams: fmt.Sprintf(`{
						"textDocument": { "uri": "%s/main.tf" },
						"range": {
							"start": { "line": 0, "character": 0 },
							"end": { "line": 1, "character": 0 }
						},
						"context": { "diagnostics": [], "only": ["source.formatAll.terraform"] }
					}`, tmpDir.URI)},
			want: fmt.Sprintf(`{
				"jsonrpc": "2.0",
				"id": 3,
				"result": [
					{
						"title":"Format Document",
						"kind":"source.formatAll.terraform",
						"edit":{
							"changes": {
								"%s/main.tf": [
									{
										"range": {
											"start": {
												"line": 0,
												"character": 0
											},
											"end": {
												"line": 1,
												"character": 0
											}
										},
										"newText": "provider \"test\" {\n"
									},
									{
										"range": {
											"start": { "line": 2, "character": 0 },
											"end": { "line": 3, "character": 0 }
										},
										"newText": "}\n"
									}
								]
							}
						}
					}
				]
				}`, tmpDir.URI),
		},
		{
			name: "source.fixAll and source.formatAll.terraform code action requested",
			request: &langserver.CallRequest{
				Method: "textDocument/codeAction",
				ReqParams: fmt.Sprintf(`{
						"textDocument": { "uri": "%s/main.tf" },
						"range": {
							"start": { "line": 0, "character": 0 },
							"end": { "line": 1, "character": 0 }
						},
						"context": { "diagnostics": [], "only": ["source.fixAll", "source.formatAll.terraform"] }
					}`, tmpDir.URI),
			},
			want: fmt.Sprintf(`{
				"jsonrpc": "2.0",
				"id": 3,
				"result": [
					{
						"title": "Format Document",
						"kind": "source.formatAll.terraform",
						"edit": {
							"changes": {
								"%s/main.tf": [
									{
										"range": {
											"start": { "line": 0, "character": 0 },
											"end": { "line": 1, "character": 0 }
										},
										"newText": "provider \"test\" {\n"
									},
									{
										"range": {
											"start": { "line": 2, "character": 0 },
											"end": { "line": 3, "character": 0 }
										},
										"newText": "}\n"
									}
								]
							}
						}
					}
				]
			}`, tmpDir.URI),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss, err := state.NewStateStore()
			if err != nil {
				t.Fatal(err)
			}

			ls := langserver.NewLangServerMock(t, NewMockSession(&MockSessionInput{
				StateStore: ss,
			}))
			stop := ls.Start(t)
			defer stop()

			ls.Call(t, &langserver.CallRequest{
				Method: "initialize",
				ReqParams: fmt.Sprintf(`{
					"capabilities": {},
					"rootUri": %q,
					"processId": 123456
			}`, tmpDir.URI)})

			ls.Notify(t, &langserver.CallRequest{
				Method:    "initialized",
				ReqParams: "{}",
			})
			ls.Call(t, &langserver.CallRequest{
				Method: "textDocument/didOpen",
				ReqParams: fmt.Sprintf(`{
				"textDocument": {
					"version": 0,
					"languageId": "terraform",
					"text": "provider  \"test\"   {\n\n      }\n",
					"uri": "%s/main.tf"
				}
			}`, tmpDir.URI)})
			waitForAllJobs(t, ss)

			ls.CallAndExpectResponse(t, tt.request, tt.want)
		})
	}
}
