# Retab VS Code Extension

A VS Code extension that integrates the retab formatter for Protocol Buffers, HCL, Terraform, Dart, and Swift files.

## Prerequisites

-   VS Code 1.74.0 or higher

## Features

Provides formatting support for:

-   Protocol Buffers (.proto)
-   HashiCorp Configuration Language (.hcl)
-   Terraform (.tf, .tfvars)
-   Dart (.dart)
-   Swift (.swift)

## Basic Usage

1. Install the extension
2. Open a supported file
3. Use VS Code's format command:
    - Command Palette: `Format Document`
    - Keyboard Shortcut: `Shift + Alt + F` (Windows/Linux) or `Shift + Option + F` (macOS)
    - Right-click menu: `Format Document`

## Workspace Configuration

To use the extension, you need to configure the workspace to use the extension as the default formatter for the supported file types.

```json
{
	"[proto]": {
		"editor.defaultFormatter": "walteh.retab-vscode"
	},
	"[hcl]": {
		"editor.defaultFormatter": "walteh.retab-vscode"
	},
	"[terraform]": {
		"editor.defaultFormatter": "walteh.retab-vscode"
	},
	"[dart]": {
		"editor.defaultFormatter": "walteh.retab-vscode"
	},
	"[swift]": {
		"editor.defaultFormatter": "walteh.retab-vscode"
	}
}
```

You may also need to enable format on save in the workspace settings:

```json
{
	"editor.formatOnSave": true
}
```

The extension will automatically detect the file type and apply the appropriate formatter.

## Configuration

The extension uses retab's configuration from your project's `.editorconfig` file. See the [retab documentation](https://github.com/walteh/retab) for more details.

## License

Apache 2.0 - See [LICENSE](../LICENSE) for more information.
