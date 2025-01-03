# Retab VS Code Extension

A VS Code extension that integrates the retab formatter for Protocol Buffers, HCL, Terraform, and Dart files.

## Prerequisites

- VS Code 1.74.0 or higher
- retab CLI installed (`go install github.com/walteh/retab/v2/cmd/retab@latest`)

## Features

Provides formatting support for:

- Protocol Buffers (.proto)
- HashiCorp Configuration Language (.hcl)
- Terraform (.tf, .tfvars)
- Dart (.dart)

## Usage

1. Install the extension
2. Open a supported file
3. Use VS Code's format command:
   - Command Palette: `Format Document`
   - Keyboard Shortcut: `Shift + Alt + F` (Windows/Linux) or `Shift + Option + F` (macOS)
   - Right-click menu: `Format Document`

The extension will automatically detect the file type and apply the appropriate formatter.

## Configuration

The extension uses retab's configuration from your project's `.editorconfig` file. See the [retab documentation](https://github.com/walteh/retab) for more details.

## Development

1. Clone the repository
2. Run `npm install` in the `code-extension` directory
3. Open the `code-extension` directory in VS Code
4. Press F5 to start debugging

## License

Apache 2.0 - See [LICENSE](../LICENSE) for more information.
