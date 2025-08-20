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
	BuildTime   string
	BuildHash   string
	BuildArch   string
	BuildTag    string
	BuildModule string
)

func main() {
	lines := strings.Split(string(gomod), "\n")
	for _, l := range lines {
		if strings.HasPrefix(l, "module ") {
			BuildModule = strings.TrimPrefix(l, "module ")
			break
		}
	}

	if err := cmd.Execute(BuildModule, BuildTag); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
