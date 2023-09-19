
BRANCH = "main"

func "leggo" {
	params = [abc, def]
	result = abc + def
}

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
				runson = "ubuntu-latest"
				steps = [
					{
						name = "Checkout"
						uses = "actions/checkout@v2"
						with = {
							fetch-depth = leggo(1, 2)
						}
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
