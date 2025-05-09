package swiftfmt_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/walteh/retab/v2/gen/mocks/pkg/formatmock"
	"github.com/walteh/retab/v2/pkg/diff"
	"github.com/walteh/retab/v2/pkg/formatters/cmdfmt"
	"github.com/walteh/retab/v2/pkg/formatters/swiftfmt"
)

func TestSwiftIntegration(t *testing.T) {
	tests := []struct {
		name       string
		useTabs    bool
		indentSize int
		src        []byte
		expected   []byte
	}{
		{
			name:       "basic formatting",
			useTabs:    false,
			indentSize: 4,
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
			name:       "complex swift code",
			useTabs:    false,
			indentSize: 4,
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
			cfg.EXPECT().IndentSize().Return(tt.indentSize).Maybe()

			ctx = zerolog.New(zerolog.NewTestWriter(t)).WithContext(ctx)

			result, err := swiftfmt.NewSwiftCmdFormatter(cmdfmt.WithUseDocker(true)).Format(ctx, cfg, bytes.NewReader(tt.src))

			require.NoError(t, err)
			diff.Require(t).Got(result).Want(tt.expected).Equals()
		})
	}
}
