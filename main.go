package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

var (
	entry = flag.String("entry", "", "Main entry point of JavaScript to bundle.")
)

func main() {
	flag.Parse()

	if *entry == "" {
		fmt.Fprintf(os.Stderr, "entry flag is required\n")
		os.Exit(1)
	}

	c := NewCompiler()

	entry, err := c.Load(context.Background(), *entry)
	if err != nil {
		panic(err)
	}

	if err := c.BundleModule(context.Background(), entry, os.Stdout); err != nil {
		panic(err)
	}
}
