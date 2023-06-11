# `tftab` 🚀✨

Welcome to a new era, where we have broken the Terraform space-time continuum.

> **TLDR:** `tftab` calls `terraform fmt` and replaces preceding spaces with tabs.

In the world of HashiCorp Configuration Language (HCL), Terraform has been telling us that spaces are the only game in town. They've got us thinking that tabs and Terraform go together like oil and water. Well, we're here to tell you that we've had enough! Enter TFTab - the renegade rebel of Terraform formatting, breaking the monotony by advocating for tabs over spaces.

We've all been there, pouring over the Terraform documentation, stumbling upon the line "HCL does not allow tabs!" But in the immortal words of Captain Barbossa, we found that to be more of a guideline than an actual rule. Turns out, Terraform HCL compiles just fine with tabs. Surprise! 😲

## 🤔 What is `tftab`?

Simply put, `tftab` is an audacious wrapper for `terraform fmt` that rebels against the status quo by using tabs instead of spaces. The real magic happens when we leverage the power of `terraform fmt`. We get the same alignment-based formatting that we know and love, but with the readability that only tabs can provide.

## 🚀 Why Use `tftab`?

Imagine reading a book where all the words are crammed together without spaces. Sounds pretty hard to read, right? That's how we feel when we see Terraform code indented with two spaces. It's just not enough. TFTab gives your code the breathing room it deserves, making it easier on the eyes and much more manageable.

And of course, there's the principle of the matter. Why should spaces hog all the limelight while tabs sit in the shadow? It's time for tabs to shine!

```hcl
# Spaces 😞
resource "aws_instance" "example" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  tags = {
    Name = "example-instance"
  }
}

# Tabs 😄
resource "aws_instance" "example" {
	ami			 = "ami-0c55b159cbfafe1f0"
	instance_type = "t2.micro"

	tags = {
		Name = "example-instance"
	}
}
```

## 🛠️ How to Install `tftab`?

Installation is a breeze with Homebrew. Just tap into the wisdom of nuggxyz, and install the formula for TFTab. Here's how you do it:

```bash
brew tap nuggxyz/tap
brew install tftab
```

And just like that, you're ready to embrace the tab life in your Terraform files.

## 🎉 Closing Thoughts

In the immortal words of a wise person we just made up, "Tabs and spaces are like the Yin and Yang of coding - different, but each with their own role to play." So why not add a little Yin (or Yang, depending on how you look at it) to your Terraform configurations? Give TFTab a spin and let those tabs strut their stuff in your `.tf` files!

So, get ready to rock the boat, challenge the norm, and introduce tabs into your Terraform files. Because, why should spaces have all the fun?

Remember, the next time someone tells you that you can't use tabs in
