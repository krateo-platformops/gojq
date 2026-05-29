// gojq - Go implementation of jq
package main

import (
	"os"

	"github.com/braghettos/gojq/cli"
)

func main() {
	os.Exit(cli.Run())
}
