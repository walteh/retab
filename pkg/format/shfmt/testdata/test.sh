#!/bin/bash
# This is a test shell script with poor formatting

# Bad indentation
if [ "$1" = "test" ]; then
    echo "This is a test"
    for i in 1 2 3; do
        echo $i
    done
fi

# Bad spacing around redirects
cat file.txt >output.txt

# Function with bad braces placement
function bad_function() {
    echo "This function has bad formatting"
}

# Binary operators on the same line
[ -f /etc/passwd ] && echo "Password file exists" || echo "No password file found"

# Case statement with poor indentation
case "$2" in
"foo")
    echo "Foo selected"
    ;;
"bar")
    echo "Bar selected"
    ;;
*)
    echo "Unknown option"
    ;;
esac
