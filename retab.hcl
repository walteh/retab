
BRANCH = "main"

file "default.yaml" {
	dir    = "./.github/workflows"
	schema = "https://raw.githubusercontent.com/SchemaStore/schemastore/master/src/schemas/json/github-workflow.json"
	data = {
		name = "test"
		on = {
			push = {
				branches = [BRANCH]
			}
		}
		jobs = {
			build = {
				runs-on = "ubuntu-latest"
				steps = [
					{
						name = "Checkout"
						uses = "actions/checkout@v2"
					},
					{
						name = "Run tests"
						run  = <<SHELL
							echo "Hello world"
						SHELL
					},
				]
			}
		}
	}
}
