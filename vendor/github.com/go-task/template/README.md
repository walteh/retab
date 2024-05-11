# go-task/template

This is a forked version of Golang's standard `text/template` package. It is
designed to be a drop-in replacement for the original package with some
additional features. This package was created by the Task project to fulfil
their project-specific needs. These features are not intended to be merged into
the original package unless they are one day deemed useful enough.

## Features

- `ResolveRef` - This package function will allow the user to give a blob of
  data as they would normally to `template.Execute` and then retrieve a value
  from that blob using go-template syntax. This solves a limitation of the
  public API of the original package which meant that it was only ever possible
  to return a string representation of a value in a template. This function is
  also available as a method on the `Template` type.
