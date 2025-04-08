# Diff Package

The `diff` package provides utilities for generating, comparing, and visualizing differences between expected and actual values in various formats. This package is designed to be used in testing scenarios but can be used in any context where difference visualization is needed.

## Features

-   Generic type comparisons using [go-cmp](https://github.com/google/go-cmp)
-   String-based diffing with unified diff format
-   Character-level diffing for detailed text comparison
-   Rich formatting with color and whitespace visualization
-   Support for various Go types including structs, maps, and primitive types
-   Reflection-based comparison for complex types

## Architecture

The package is organized into several modules:

1. **Core Diff Logic** (`diff.go`): Main interfaces and type-specific diff generation
2. **Unified Diff** (`udiff.go`): Implementation of unified diff parsing and formatting
3. **Output Enrichment** (`enrich.go`): Enhances diff output with colors and formatting
4. **Type & Value Formatting** (`convoluted.go`): Formats Go types and values for comparison
5. **Testing Utilities** (`testing.go`): Functions for using diffs in testing

## Usage Examples

### Basic Comparison

```go
package main

import (
    "fmt"
    "github.com/walteh/cloudstack-mcp/pkg/diff"
)

func main() {
    expected := "Hello, World!"
    actual := "Hello, Universe!"

    // Generate a diff between two strings
    result := diff.TypedDiff(expected, actual)

    // Print the diff
    fmt.Println(result)
}
```

### Using in Tests

```go
package mypackage_test

import (
    "testing"
    "github.com/walteh/cloudstack-mcp/pkg/diff"
)

func TestSomething(t *testing.T) {
    expected := map[string]int{"a": 1, "b": 2}
    actual := map[string]int{"a": 1, "b": 3}

    // Compare values and fail if they don't match
    diff.RequireEqual(t, expected, actual)
}
```

### Comparing Complex Types

```go
package mypackage_test

import (
    "reflect"
    "testing"
    "github.com/walteh/cloudstack-mcp/pkg/diff"
)

type MyStruct struct {
    Name string
    Age  int
    private string
}

func TestStructComparison(t *testing.T) {
    s1 := MyStruct{Name: "Alice", Age: 30, private: "secret"}
    s2 := MyStruct{Name: "Alice", Age: 31, private: "secret"}

    // Compare with unexported fields
    diff.RequireEqual(t, s1, s2, diff.WithUnexportedType[MyStruct]())
}
```

### Character-Level Diffing

```go
package main

import (
    "fmt"
    "github.com/walteh/cloudstack-mcp/pkg/diff"
)

func main() {
    str1 := "The quick brown fox jumps over the lazy dog"
    str2 := "The quick brown fox jumps over the lazy cat"

    // Get character-level differences
    result := diff.SingleLineStringDiff(str1, str2)

    fmt.Println(result)
}
```

## Extension Points

The package is designed to be extensible:

1. Create custom formatters by implementing the `DiffFormatter` interface
2. Create custom diff generators by implementing the `Differ` interface
3. Add options to modify behavior via the various `...Opts` structs

## Best Practices

1. Use `TypedDiff` for comparing values of the same type
2. Use `TypedDiffExportedOnly` when comparing structs where unexported fields should be ignored
3. Use `RequireEqual` and similar functions in tests for clean assertion code
4. Consider using `WithUnexportedType` when comparing structs with private fields in tests
5. For long diffs, the output is automatically shortened to make it more readable

## Legacy Support

The package maintains backward compatibility with older functions like:

-   `RequireUnknownTypeEqual`
-   `RequireUnknownValueEqualAsJSON`
-   `RequireKnownValueEqual`

However, it's recommended to use the newer, more consistently named functions instead.
