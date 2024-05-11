<p align="center">
<img src="./img/gopher.png" width="250" >
</p>

# retab ➡️

**Effortless Configuration Management with HCL**

`retab` streamlines your configuration workflow by transforming human-readable HCL (HashiCorp Configuration Language) files into YAML or JSON.  Enjoy the benefits of:

* **Simplified Syntax:** Write cleaner, more maintainable configurations using HCL.
* **Flexible Output:** Generate YAML or JSON output to match your project's requirements.
* **Improved Readability:**  Automatically format HCL files for optimal clarity and consistency.

## Installation

```bash
go install github.com/walteh/retab/cmd/retab
```

## Usage

![WARNING]Files must be located in a `.retab` directory and have the `.retab` extension.

1. **Write your configuration in HCL**
2. **Format:** `retab fmt`
3. **Generate:** `retab gen` (outputs YAML or JSON)

## Example

```hcl
# ./.retab/config.retab
gen "config" {
	schema = "https://example.com/schema.json" # Optional
	path   = "config.yaml"
	data = {
		server = {
			port       = 8080
			enable_ssl = true
		}
	}
}
```

Run `retab gen` to produce:

```yaml
# ./config.yaml
server:
  port: 8080
  enable_ssl: true
```

## Additional Features

* **Advanced Formatting:** Format HCL, Terraform (.tf), Protocol Buffers (.proto), and Dart (.dart) files.
* **Schema Validation:** Ensure your configurations adhere to defined schemas.

