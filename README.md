# `retab` 🚀✨

imagine a world where you don't have to fight with `yaml` anymore

retab empowers you to write your `yaml` files in `hcl` ... and format them with tabs 🤯

---

## 🛠️ install from the source

```bash
go install  github.com/walteh/retab/cmd/retab
```
---

## 3️⃣ commands, endless possibilities

```bash
retab gen
```
✅ use `hcl` to write `.yaml` and `.json` files

✅ never fight with yaml again 🎉

✅ runs all `.retab` files in the current directory, or in the `.retab` directory

```bash
retab fmt --file=<file>
```
✅ formats `.hcl` files with its standard guidelines, but with tabs as indentation

✅ also supports `.proto` files

✅ native, no runtime dependancies

- > ⚠️  `.tf` and `.dart` files are supported, but depend on the `terraform` and `dart` commands being installed on your system

---

<!-- ## supplemental commands

```bash
retab wfmt --command=<command> --file=<file>
```
✅ "wrapped `fmt`" - runs a command on your system, and formats the output with tabs

⚠️ depends on the command you want to run

terraform example: `retab wfmt --terraform --file=main.tf`

--- -->


## `retab gen`: builing `.yaml` and `.json` files

start with a `.retab` file

```hcl
# my-workflow.retab
file "deployment-pipeline.yaml" {
	dir = "./.codecatalyst/workflows"
	data = {
		Name          = "deployment-pipeline"
		RunMode       = "SUPERSEDED"
		SchemaVersion = "1.0"
		Triggers = [
			{
				Branches = ["main"]
				Type     = "PUSH"
			}
		]
		Actions = {
			Beta = Actions.Beta
		}
	}
}

Actions "Beta" {
	Actions "Deploy" {
		Identifier = "aws/cdk-deploy@v1"
		Inputs = {
			Sources = ["WorkflowSource"]
		}
	}
}
```

run `retab gen`

```yaml
# ./.github/workflows/my-workflow.yaml

# managed by retab - please do not edit manually
# join the fight against yaml @ github.com/walteh/retab

# source ./my-workflow.retab

Name: deployment-pipeline
RunMode: SUPERSEDED
SchemaVersion: 1.0
Triggers:
    - Branches:
	    - main
	Type: PUSH
Actions:
    Beta:
	  Deploy:
	    Identifier: aws/cdk-deploy@v1
	    Inputs:
			- Sources:
				 - WorkflowSource
```

see  your new `.yaml` file in the `.github/workflows` directory

---

## `retab fmt`: formatting `.hcl` files

welcome to a new era, where we have broken the YAML/HCL space-time continuum.



<!-- `retab` is the same as `terraform fmt`, but it replaces **preceding** spaces with tabs. -->

```hcl
# terraform fmt 😞
resource "aws_instance" "example" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  tags = {
    Name = "example-instance"
  }
}

# retab 😄
resource "aws_instance" "example" {
	ami           = "ami-0c55b159cbfafe1f0"
	instance_type = "t2.micro"

	tags = {
		Name = "example-instance"
	}
}
```

## 🏓 use `retab`

<!-- Using `retab` is as easy as it gets. Just replace `terraform fmt` with `retab` in your workflow, and you're all set. Here's how you do it: -->

```bash
# Before
terraform fmt my_sad_file.tf

# After
retab my_happy_file.tf
```




<!-- Installation is a breeze with Go. Here's how you do it: -->

```bash
go install  github.com/walteh/retab/cmd/retab
```

<!-- And just like that, you're ready to embrace the tab life in your HCL files. -->

## 🤯 Why `retab`?

In the world of HashiCorp Configuration Language (HCL), Terraform has been telling us that spaces are the only game in town. They've got us thinking that tabs and Terraform go together like oil and water. Well, we're here to tell you that we've had enough! Enter `retab` - the renegade rebel of Terraform formatting, breaking the monotony by advocating for tabs over spaces.

We've all been there, pouring over the Terraform documentation, stumbling upon the line "HCL does not allow tabs!" But in the immortal words of Captain Barbossa, we found that to be more of a guideline than an actual rule. Turns out, Terraform HCL compiles just fine with tabs. Surprise! 😲

<!-- ## 🤔 What is `retab`?

Simply put, `retab` is an audacious wrapper for `terraform fmt` that rebels against the status quo by using tabs instead of spaces. The real magic happens when we leverage the power of `terraform fmt`. **We get the same alignment-based formatting that we know and love, but with the readability that only tabs can provide.**

## 🚀 Why Use `retab`?

Imagine reading a book where all the words are crammed together without spaces. Sounds pretty hard to read, right? That's how we feel when we see Terraform code indented with two spaces. It's just not enough. `retab` gives your code the breathing room it deserves, making it easier on the eyes and much more manageable.

And of course, there's the principle of the matter. Why should spaces hog all the limelight while tabs sit in the shadow? It's time for tabs to shine! -->


## 🤓 How Does `retab` Work?

`retab` is a simple wrapper for `terraform fmt` that replaces preceding spaces with tabs. It's a one-trick pony, but it's a pretty neat trick. Here's how it works:

1. `retab` calls `terraform fmt` to format the Terraform files.
2. `retab` then replaces all preceding spaces with tabs using `sed`.

That's it! It's that simple. And the best part is that you don't have to change your workflow at all. Just use `retab` instead of `terraform fmt`, and you're good to go.


## 🎉 Closing Thoughts

In the grand scheme of things, the choice between spaces and tabs might seem trivial. But when you're deep in the trenches of Terraform code, every little bit of readability and organization counts. With `retab`, you're not just choosing tabs over spaces - you're choosing a new way to experience Terraform.

As we step into this bold new era of Terraform formatting, we believe that a little rebellion can lead to a world of difference. With `retab`, we challenge the conventions and celebrate the beauty of diversity in code formatting. Because if coding teaches us anything, it's that there's always more than one way to solve a problem.

So why stick to the status quo when you can have tabs that not only break the monotony but also bring a refreshing change to your Terraform codebase? Use `retab`, embrace the change, and let your .tf files revel in the joy of tabs!

In the end, whether you're a spaces loyalist or a tabs enthusiast, the goal remains the same – to write Terraform code that's clean, organized, and easy to understand. And if tabs can help you do that better, then why not give it a shot?

So, go ahead, take the leap of faith, and let the tabs take center stage in your Terraform files. Because in the world of code, there's always room for a little revolution.
