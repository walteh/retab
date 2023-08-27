package bufwrite_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walteh/tftab/gen/mockery"
	"github.com/walteh/tftab/pkg/bufwrite"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name       string
		useTabs    bool
		indentSize int
		src        []byte
		expected   []byte
	}{
		{
			name:       "Use Tabs with IndentSize 1",
			useTabs:    true,
			indentSize: 1,
			src: []byte(`
syntax = "proto3";
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
  string payload =     6     ;
}

message EnvironmentOptionsRequest {
  string service = 1;
  oneof request {
    RequestQuickInfoResponse request_quick_info = 2;
  }
  }

service OgWebServerService {
  rpc EnvironmentOptions(EnvironmentOptionsRequest) returns (EnvironmentOptionsResponse) {}
}
`),
			expected: []byte(`syntax = "proto3";
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
	}
}

service OgWebServerService {
	rpc EnvironmentOptions(EnvironmentOptionsRequest) returns (EnvironmentOptionsResponse) {}
}
`),
		},
		{
			name:       "Use Spaces with IndentSize 4",
			useTabs:    false,
			indentSize: 4,
			src: []byte(`
syntax = "proto3";
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
  string payload =     6     ;
}

message EnvironmentOptionsRequest {
  string service = 1;
  oneof request {
    RequestQuickInfoResponse request_quick_info = 2;
  }
}

service OgWebServerService {
  rpc EnvironmentOptions(EnvironmentOptionsRequest) returns (EnvironmentOptionsResponse) {}
}`),
			expected: []byte(`syntax = "proto3";
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
    }
}

service OgWebServerService {
    rpc EnvironmentOptions(EnvironmentOptionsRequest) returns (EnvironmentOptionsResponse) {}
}
`),
		},
		// Add more test cases here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()

			cfg := &mockery.MockProvider_configuration{}
			cfg.EXPECT().UseTabs().Return(tt.useTabs)
			cfg.EXPECT().IndentSize().Return(tt.indentSize)

			// Call the Format function with the provided configuration and source
			result, err := bufwrite.NewBufFormatter().Format(ctx, cfg, bytes.NewReader(tt.src))

			// Check for errors
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Read the result into a buffer
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(result)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Compare the result with the expected outcome
			assert.Equal(t, string(tt.expected), buf.String(), " source does not match expected output")
		})
	}
}
