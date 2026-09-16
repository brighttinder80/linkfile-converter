package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintf(os.Stderr, `dotconv converts dotfile symlink definitions between two formats:

  linkfile   a plain "target = source" manifest, one mapping per line
  script     a shell script of "ln -s" commands

Usage:
  dotconv to-script   [--lenient] [-o out.sh] <linkfile>
  dotconv to-linkfile [--lenient] [-o out.linkfile] <script.sh>
  dotconv validate    [--lenient] <linkfile>

validate parses a linkfile and reports errors without producing any
output file. It's meant for a pre-commit hook or CI check on a
dotfiles repo that keeps its linkfile as the source of truth.

By default both directions are strict: anything that doesn't parse
cleanly, or looks unsafe (a symlink target that isn't absolute or
~-rooted, a source that escapes the repo with ".."), stops the
conversion with an error. Pass --lenient to skip the offending lines
instead, with a warning on stderr for each one.
`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	switch cmd {
	case "to-script", "to-linkfile", "validate":
	default:
		usage()
		os.Exit(2)
	}

	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	lenient := fs.Bool("lenient", false, "skip malformed or unsafe lines instead of failing")
	var output *string
	if cmd != "validate" {
		output = fs.String("o", "", "output file (default: stdout)")
	}
	fs.Parse(os.Args[2:])

	if fs.NArg() != 1 {
		usage()
		os.Exit(2)
	}

	in, err := os.Open(fs.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "dotconv:", err)
		os.Exit(1)
	}
	defer in.Close()

	if cmd == "validate" {
		entries, err := ParseLinkfile(in, *lenient, os.Stderr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "dotconv:", err)
			os.Exit(1)
		}
		fmt.Printf("ok: %d entries\n", len(entries))
		return
	}

	out := os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			fmt.Fprintln(os.Stderr, "dotconv:", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}

	switch cmd {
	case "to-script":
		entries, err := ParseLinkfile(in, *lenient, os.Stderr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "dotconv:", err)
			os.Exit(1)
		}
		if err := GenerateScript(out, entries); err != nil {
			fmt.Fprintln(os.Stderr, "dotconv:", err)
			os.Exit(1)
		}
	case "to-linkfile":
		entries, err := ExtractFromScript(in, *lenient, os.Stderr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "dotconv:", err)
			os.Exit(1)
		}
		if err := WriteLinkfile(out, entries); err != nil {
			fmt.Fprintln(os.Stderr, "dotconv:", err)
			os.Exit(1)
		}
	}
}
