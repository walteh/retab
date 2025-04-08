package protofmt_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/walteh/retab/v2/gen/mocks/pkg/formatmock"
	"github.com/walteh/retab/v2/pkg/format"
	"github.com/walteh/retab/v2/pkg/format/protofmt"
)

func formatProto(ctx context.Context, cfg format.Configuration, src []byte) (string, error) {
	formatter := protofmt.NewFormatter()
	reader, err := formatter.Format(ctx, cfg, bytes.NewReader(src))
	if err != nil {
		return "", err
	}

	result, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

type formatTest struct {
	name     string
	useTabs  bool
	src      string
	expected string
}

func visualizeWhitespace(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, " ", "·"), "\t", "→")
}

func TestAlignmentCases(t *testing.T) {
	tests := []formatTest{
		{
			name:    "Simple Field Alignment",
			useTabs: true,
			src: `message Test {
  string short = 1;
  string very_long_field = 2;
  int32 medium = 3;
}`,
			expected: `message Test {
	string short           = 1;
	string very_long_field = 2;
	int32  medium          = 3;
}`,
		},
		{
			name:    "Mixed Type Field Alignment",
			useTabs: true,
			src: `message MixedTypes {
  string name = 1;
  repeated int32 numbers = 2;
  map<string, bool> settings = 3;
  optional bytes data = 4;
}`,
			expected: `message MixedTypes {
	string            name     = 1;
	repeated int32    numbers  = 2;
	map<string, bool> settings = 3;
	optional bytes    data     = 4;
}`,
		},
		{
			name:    "Enum Alignment",
			useTabs: true,
			src: `enum Status {
  STATUS_UNSPECIFIED = 0;
  PENDING = 1;
  IN_PROGRESS = 2;
  COMPLETED = 3;
}`,
			expected: `enum Status {
	STATUS_UNSPECIFIED = 0;
	PENDING            = 1;
	IN_PROGRESS        = 2;
	COMPLETED          = 3;
}`,
		},
		{
			name:    "Full Example",
			useTabs: true,
			src: `syntax = "proto3";
package webauthn;

option go_package = "og/gen/buf/go/proto/server";

message EnvironmentOptionsResponse {
repeated string environment_options = 1;
}

message RequestQuickInfoResponse {
string id = 1;
string service = 2;
string path = 3;
string true_path = 4;
string method = 5;
string payload = 6;
}

message EnvironmentOptionsRequest {
string service = 1;
oneof request {
RequestQuickInfoResponse request_quick_info = 2;
RequestQuickInfoResponse request_quick_info_other = 3;
}
}`,
			expected: `syntax = "proto3";
package webauthn;

option go_package = "og/gen/buf/go/proto/server";

message EnvironmentOptionsResponse {
	repeated string environment_options = 1;
}

message RequestQuickInfoResponse {
	string id        = 1;
	string service   = 2;
	string path      = 3;
	string true_path = 4;
	string method    = 5;
	string payload   = 6;
}

message EnvironmentOptionsRequest {
	string service = 1;
	oneof request {
		RequestQuickInfoResponse request_quick_info       = 2;
		RequestQuickInfoResponse request_quick_info_other = 3;
	}
}`,
		},
		{
			name:    "Service Options",
			useTabs: true,
			src: `service ScriptService {
	option (tools.v1.goreleaser_template_lambda) = "./backend/cmd/lambdas/script.v1.ScriptService";
	option (tools.v1.env) = "DYNAMODB_TABLE_NAME";
	option (tools.v1.env) = "DYNAMODB_TABLE_REGION";
	option (tools.v1.env) = "JWT_ISSUER";
	option (tools.v1.env) = "JWT_AUDIENCE";
	option (tools.v1.env) = "JWT_LEEWAY";

	rpc UpdateScript(UpdateScriptRequest) returns (UpdateScriptResponse) {
		option (tools.v1.aws_action) = "dynamodb:UpdateItem";
		option (tools.v1.gg) = "dynamodb:PutItem";
	}
	rpc ListScripts(ListScriptsRequest) returns (ListScriptsResponse) {
		option (tools.v1.aws_action) = "dynamodb:Query";
		option (tools.v1.zz) = "dynamodb:Scan";
	}
	rpc ExecuteScript(ExecuteScriptRequest) returns (ExecuteScriptResponse) {
		option (tools.v1.aws_action) = "dynamodb:GetItem";
		option (tools.v1.xx) = "lambda:InvokeFunction";
	}
}
`,
			expected: `service ScriptService {
	option (tools.v1.goreleaser_template_lambda) = "./backend/cmd/lambdas/script.v1.ScriptService";
	option (tools.v1.env)                        = "DYNAMODB_TABLE_NAME";
	option (tools.v1.env)                        = "DYNAMODB_TABLE_REGION";
	option (tools.v1.env)                        = "JWT_ISSUER";
	option (tools.v1.env)                        = "JWT_AUDIENCE";
	option (tools.v1.env)                        = "JWT_LEEWAY";

	rpc UpdateScript(UpdateScriptRequest) returns (UpdateScriptResponse) {
		option (tools.v1.aws_action) = "dynamodb:UpdateItem";
		option (tools.v1.gg)         = "dynamodb:PutItem";
	}

	rpc ListScripts(ListScriptsRequest) returns (ListScriptsResponse) {
		option (tools.v1.aws_action) = "dynamodb:Query";
		option (tools.v1.zz)         = "dynamodb:Scan";
	}

	rpc ExecuteScript(ExecuteScriptRequest) returns (ExecuteScriptResponse) {
		option (tools.v1.aws_action) = "dynamodb:GetItem";
		option (tools.v1.xx)         = "lambda:InvokeFunction";
	}
}`,
		},

		{
			name:    "Compact Options",
			useTabs: true,
			src: `message ComplexOptions {
	option (custom.option)          = true;
	option (another.complex.option) = "value";
	option allow_alias              = true;


	string name = 1 [
			(custom.field) = "value", // should be aligned
			deprecated = true // should be aligned
	];
	int32  id = 2 [
			(validate.rules).int32.gt = 0
	];
}
`,
			expected: `message ComplexOptions {
	option (custom.option)          = true;
	option (another.complex.option) = "value";
	option allow_alias              = true;

	string name = 1 [
		(custom.field) = "value",  // should be aligned
		deprecated     = true      // should be aligned
	];
	int32  id   = 2 [
		(validate.rules).int32.gt = 0
	];
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg := formatmock.NewMockConfiguration(t)
			cfg.EXPECT().UseTabs().Return(tt.useTabs).Maybe()
			cfg.EXPECT().IndentSize().Return(1).Maybe()

			// add a newline at the end of the src
			if !strings.HasSuffix(tt.src, "\n") {
				tt.src = tt.src + "\n"
			}
			if !strings.HasSuffix(tt.expected, "\n") {
				tt.expected = tt.expected + "\n"
			}

			formatted, err := formatProto(ctx, cfg, []byte(tt.src))
			if err != nil {
				t.Fatalf("Format returned error: %v", err)
			}

			if got := formatted; got != tt.expected {
				t.Errorf("Format returned incorrect result.\nExpected (with whitespace):\n%s\nGot (with whitespace):\n%s",
					visualizeWhitespace(tt.expected),
					visualizeWhitespace(got))

				// Show line by line comparison for easier debugging
				expectedLines := strings.Split(tt.expected, "\n")
				gotLines := strings.Split(got, "\n")

				minLen := len(expectedLines)
				if len(gotLines) < minLen {
					minLen = len(gotLines)
				}

				for i := 0; i < minLen; i++ {
					if expectedLines[i] != gotLines[i] {
						t.Errorf("Line %d difference:\nExpected: %s\nGot:      %s",
							i+1,
							visualizeWhitespace(expectedLines[i]),
							visualizeWhitespace(gotLines[i]))
					}
				}

				if len(expectedLines) != len(gotLines) {
					t.Errorf("Line count mismatch: expected %d, got %d", len(expectedLines), len(gotLines))
				}
			}
		})
	}
}

func TestBasicFieldAlignment(t *testing.T) {
	input := `message Test {
	string short = 1;
	string very_long_field = 2;
	int32 medium = 3;
}`

	expected := "message Test {\n" +
		"\tstring short           = 1;\n" +
		"\tstring very_long_field = 2;\n" +
		"\tint32  medium          = 3;\n" +
		"}\n"

	cfg := formatmock.NewMockConfiguration(t)
	cfg.EXPECT().UseTabs().Return(true).Maybe()
	cfg.EXPECT().IndentSize().Return(1).Maybe()

	formatted, err := formatProto(context.Background(), cfg, []byte(input))
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	if got := formatted; got != expected {
		t.Errorf("Format returned incorrect result.\nExpected (with whitespace):\n%s\nGot (with whitespace):\n%s",
			visualizeWhitespace(expected),
			visualizeWhitespace(got))

		// Show line by line comparison
		expectedLines := strings.Split(expected, "\n")
		gotLines := strings.Split(got, "\n")

		minLen := len(expectedLines)
		if len(gotLines) < minLen {
			minLen = len(gotLines)
		}

		for i := 0; i < minLen; i++ {
			if expectedLines[i] != gotLines[i] {
				t.Errorf("Line %d difference:\nExpected: %s\nGot:      %s",
					i+1,
					visualizeWhitespace(expectedLines[i]),
					visualizeWhitespace(gotLines[i]))
			}
		}

		t.Errorf("Expected line lengths: %v", []int{
			len("\tstring short           = 1;"),
			len("\tstring very_long_field = 2;"),
			len("\tint32  medium          = 3;"),
		})
		t.Errorf("Got line lengths: %v", []int{
			len(strings.Split(got, "\n")[1]),
			len(strings.Split(got, "\n")[2]),
			len(strings.Split(got, "\n")[3]),
		})
	}
}
