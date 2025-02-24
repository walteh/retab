# retab

A powerful multi-language code formatter that emphasizes tabs-first formatting, with native support for Protocol Buffers and HCL files, plus additional support for external formatters like Terraform, Dart, and Swift.

## Installation

```bash
go install github.com/walteh/retab/v2/cmd/retab@latest
```

## Features

-   **Native Formatters:**

    -   Protocol Buffers (.proto files)
    -   HashiCorp Configuration Language (HCL)

-   **External Formatters:**

    -   Terraform (requires `terraform` CLI)
    -   Dart (requires `dart` CLI)
    -   Swift (requires `swift-format`)

-   **Tabs-First Approach:** While the formatter respects your `.editorconfig` settings, it's designed with tabs in mind for better accessibility and consistent indentation.

## Usage

Format a file using the `fmt` command:

```bash
# Auto-detect formatter based on file extension
retab fmt myfile.proto

# Explicitly specify formatter
retab fmt myfile.proto --formatter=proto
retab fmt myfile.hcl --formatter=hcl
retab fmt myfile.tf --formatter=tf
retab fmt myfile.dart --formatter=dart
retab fmt myfile.swift --formatter=swift
```

## Examples

### Protocol Buffers

```protobuf
// Before formatting
service  MyService{rpc   MyMethod(MyRequest)   returns(MyResponse);}

// After formatting (with default tab indentation)
service MyService {
	rpc MyMethod(MyRequest) returns (MyResponse);
}
```

### HCL

```hcl
// Before formatting
resource "aws_instance" "example" {ami="ami-123456"
instance_type="t2.micro"
  tags={Name="example"}}

// After formatting (with default tab indentation)
resource "aws_instance" "example" {
	ami           = "ami-123456"
	instance_type = "t2.micro"
	tags = {
		Name = "example"
	}
}
```

### Swift

```swift
// Before formatting
struct ContentView{
var body:some View{
Text("Hello, world!")
.padding()
}}

// After formatting (with default indentation)
struct ContentView {
    var body: some View {
        Text("Hello, world!")
            .padding()
    }
}
```

## Configuration

retab uses `.editorconfig` for configuration. While designed with tabs in mind, it respects your project's settings. Here's a sample `.editorconfig`:

```ini
[*]
# common settings supported
indent_style = tab   # 'tab' or 'space'
indent_size = 4     # Size of indentation

# custom settings supported
trim_multiple_empty_lines = true  # Remove multiple blank lines
one_bracket_per_line = true  # Force brackets onto new lines
```

If no `.editorconfig` is found, it defaults to:

-   Tabs for indentation (recommended)
-   Tab size of 4
-   Trim multiple empty lines enabled
-   One bracket per line enabled

### Swift Formatting Note

When using Swift formatting with EditorConfig, indentation settings will only work correctly if your `swift-format` configuration has `spaces=2` set as the indentation (which is the default). If you need different indentation settings, you'll need to modify your `swift-format` configuration file.

### Why Tabs?

We believe in tabs-first formatting because:

-   Better accessibility for developers using screen readers
-   Allows each developer to set their preferred indentation width
-   Smaller file sizes
-   Clear and unambiguous indentation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

-   Built with [protocompile](https://github.com/bufbuild/protocompile) for Protocol Buffer formatting
-   Uses [editorconfig-core-go](https://github.com/editorconfig/editorconfig-core-go) for configuration
