package cmdfmt_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/walteh/retab/v2/gen/mocks/pkg/formatmock"
	"github.com/walteh/retab/v2/pkg/format/cmdfmt"
)

func TestSwiftIntegration(t *testing.T) {
	tests := []struct {
		name                   string
		useTabs                bool
		indentSize             int
		trimMultipleEmptyLines bool
		src                    []byte
		expected               []byte
	}{
		{
			name:                   "basic formatting",
			useTabs:                false,
			indentSize:             4,
			trimMultipleEmptyLines: true,
			src: []byte(`
struct ContentView {
var body: some View {
Text("Hello, world!")
.padding()
}
}
`),
			expected: []byte(`struct ContentView {
    var body: some View {
        Text("Hello, world!")
            .padding()
    }
}
`),
		},
		{
			name:                   "complex swift code",
			useTabs:                false,
			indentSize:             4,
			trimMultipleEmptyLines: true,
			src: []byte(`
import SwiftUI

@main
struct MyApp: App {
var body: some Scene {
WindowGroup {
ContentView()
}
}
}

struct ContentView: View {
@State private var isPresented = false
@State private var counter = 0

var body: some View {
VStack {
Text("Counter: \(counter)")
.font(.title)
Button("Increment") {
counter += 1
}
.padding()
Button("Show Sheet") {
isPresented = true
}
}
.sheet(isPresented: $isPresented) {
Text("Modal Sheet")
}
}
}
`),
			expected: []byte(`import SwiftUI

@main
struct MyApp: App {
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}

struct ContentView: View {
    @State private var isPresented = false
    @State private var counter = 0

    var body: some View {
        VStack {
            Text("Counter: \(counter)")
                .font(.title)
            Button("Increment") {
                counter += 1
            }
            .padding()
            Button("Show Sheet") {
                isPresented = true
            }
        }
        .sheet(isPresented: $isPresented) {
            Text("Modal Sheet")
        }
    }
}
`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			cfg := formatmock.NewMockConfiguration(t)
			cfg.EXPECT().UseTabs().Return(tt.useTabs)
			cfg.EXPECT().IndentSize().Return(tt.indentSize)
			cfg.EXPECT().TrimMultipleEmptyLines().Return(tt.trimMultipleEmptyLines)

			result, err := cmdfmt.NewSwiftFormatter(
				// --interactive allows us to read from stdin
				// --quiet suppresses the pull information in case the image is not available locally
				"docker", "run", "--interactive", "--quiet", "swift:latest", "swift-format",
			).Format(ctx, cfg, bytes.NewReader(tt.src))

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(result)
			if err != nil {
				t.Fatalf("Unexpected error reading result: %v", err)
			}

			assert.Equal(t, string(tt.expected), buf.String(), "formatted source does not match expected output")
		})
	}
}
