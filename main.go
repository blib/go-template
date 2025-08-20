package main

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/blib/go-template/cmd"
)

//go:embed go.mod
var gomod []byte

var (
	commit  string
	date    string //nolint:unused // build date
	builtBy string //nolint:unused // builder
	module  string
)

func main() {
	lines := strings.Split(string(gomod), "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "module ") {
			module = strings.TrimPrefix(l, "module ")
			break
		}
	}

	if err := cmd.Execute(module, commit); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
