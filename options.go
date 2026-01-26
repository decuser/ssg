// options.go
package main

import (
	"flag"
	"fmt"
	"os"
)

const appName = "ssg"

var version = "dev"

type Options struct {
	Version bool
	Debug   bool
	Draft   bool
	Output  string
	Posts   string
}

func parseOptions() Options {
	showVersion := flag.Bool("v", false, "show version and exit")
	debug := flag.Bool("d", false, "debug output")
	draft := flag.Bool("draft", false, "include draft posts")
	output := flag.String("o", "", "output directory (required)")
	posts := flag.String("p", "_posts", "posts directory")
	help := flag.Bool("h", false, "show help")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "ssg - static site generator")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "usage: ssg -o <output> [options]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "options:")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("%s %s\n", appName, version)
		os.Exit(0)
	}

	if *output == "" {
		fmt.Fprintln(os.Stderr, "error: -o output directory required")
		flag.Usage()
		os.Exit(1)
	}

	return Options{
		Version: *showVersion,
		Debug:   *debug,
		Draft:   *draft,
		Output:  *output,
		Posts:   *posts,
	}
}
