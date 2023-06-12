# `tftab` 🚀✨

Welcome to a new era, where we have broken the Terraform space-time continuum.

---

✅ `=` sign alignment

**AND**

✅ tab based indentation

---

<!-- `tftab` is the same as `terraform fmt`, but it replaces **preceding** spaces with tabs. -->


```hcl
# terraform fmt 😞
resource "aws_instance" "example" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  tags = {
    Name = "example-instance"
  }
}

# tftab 😄
resource "aws_instance" "example" {
	ami           = "ami-0c55b159cbfafe1f0"
	instance_type = "t2.micro"

	tags = {
		Name = "example-instance"
	}
}
```

## 🏓 use `tftab`

<!-- Using `tftab` is as easy as it gets. Just replace `terraform fmt` with `tftab` in your workflow, and you're all set. Here's how you do it: -->

```bash
# Before
terraform fmt my_sad_file.tf

# After
tftab my_happy_file.tf
```

## 🛠️ install `tftab`

### 🍺 Homebrew

<!-- Installation is a breeze with Homebrew. Just tap into the wisdom of nuggxyz, and install the formula for `tftab`. Here's how you do it: -->

```bash
brew tap nuggxyz/tap
brew install tftab
```

<!-- And just like that, you're ready to embrace the tab life in your Terraform files. -->

## 🤯 Why `tftab`?

In the world of HashiCorp Configuration Language (HCL), Terraform has been telling us that spaces are the only game in town. They've got us thinking that tabs and Terraform go together like oil and water. Well, we're here to tell you that we've had enough! Enter `tftab` - the renegade rebel of Terraform formatting, breaking the monotony by advocating for tabs over spaces.

We've all been there, pouring over the Terraform documentation, stumbling upon the line "HCL does not allow tabs!" But in the immortal words of Captain Barbossa, we found that to be more of a guideline than an actual rule. Turns out, Terraform HCL compiles just fine with tabs. Surprise! 😲

<!-- ## 🤔 What is `tftab`?

Simply put, `tftab` is an audacious wrapper for `terraform fmt` that rebels against the status quo by using tabs instead of spaces. The real magic happens when we leverage the power of `terraform fmt`. **We get the same alignment-based formatting that we know and love, but with the readability that only tabs can provide.**

## 🚀 Why Use `tftab`?

Imagine reading a book where all the words are crammed together without spaces. Sounds pretty hard to read, right? That's how we feel when we see Terraform code indented with two spaces. It's just not enough. `tftab` gives your code the breathing room it deserves, making it easier on the eyes and much more manageable.

And of course, there's the principle of the matter. Why should spaces hog all the limelight while tabs sit in the shadow? It's time for tabs to shine! -->


## 🤓 How Does `tftab` Work?

`tftab` is a simple wrapper for `terraform fmt` that replaces preceding spaces with tabs. It's a one-trick pony, but it's a pretty neat trick. Here's how it works:

1. `tftab` calls `terraform fmt` to format the Terraform files.
2. `tftab` then replaces all preceding spaces with tabs using `sed`.

That's it! It's that simple. And the best part is that you don't have to change your workflow at all. Just use `tftab` instead of `terraform fmt`, and you're good to go.


## 🎉 Closing Thoughts

In the grand scheme of things, the choice between spaces and tabs might seem trivial. But when you're deep in the trenches of Terraform code, every little bit of readability and organization counts. With `tftab`, you're not just choosing tabs over spaces - you're choosing a new way to experience Terraform.

As we step into this bold new era of Terraform formatting, we believe that a little rebellion can lead to a world of difference. With `tftab`, we challenge the conventions and celebrate the beauty of diversity in code formatting. Because if coding teaches us anything, it's that there's always more than one way to solve a problem.

So why stick to the status quo when you can have tabs that not only break the monotony but also bring a refreshing change to your Terraform codebase? Use `tftab`, embrace the change, and let your .tf files revel in the joy of tabs!

In the end, whether you're a spaces loyalist or a tabs enthusiast, the goal remains the same – to write Terraform code that's clean, organized, and easy to understand. And if tabs can help you do that better, then why not give it a shot?

So, go ahead, take the leap of faith, and let the tabs take center stage in your Terraform files. Because in the world of code, there's always room for a little revolution.
