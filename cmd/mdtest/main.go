package main

import (
	"os"

	"github.com/PeronGH/mdtest-cli/internal/cli"
)

func main() {
	code := cli.Execute(os.Args[1:], os.Stdout, os.Stderr, cli.DefaultLookPath)
	os.Exit(code)
}
