package main

import (
	"fmt"
	"os"

	"github.com/blib/go-template/cmd"
)

var BuildTime string
var BuildHash string
var BuildPlatform string

func main() {
	fmt.Fprintf(os.Stderr, "Backend v.%s build at %s for %s\n", BuildHash, BuildTime, BuildPlatform)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
