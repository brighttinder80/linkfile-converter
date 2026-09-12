package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Entry is one target-to-source mapping: Target is where the symlink
// will be created, Source is the file it should point to inside the
// dotfiles repo.
type Entry struct {
	Target string
	Source string
	Line   int
}

// ParseLinkfile reads the linkfile format:
//
//	target = source
//
// Blank lines and lines starting with # are ignored. In strict mode
// any malformed or suspicious line (missing '=', a source that
// escapes the repo with "..", a target that isn't absolute or
// "~"-rooted) aborts parsing with an error. In lenient mode such
// lines are reported on warn and kept or skipped depending on
// whether they could still be interpreted.
func ParseLinkfile(r io.Reader, lenient bool, warn io.Writer) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(r)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			if !lenient {
				return nil, fmt.Errorf("line %d: missing '=' separator: %q", lineNo, line)
			}
			fmt.Fprintf(warn, "skipping line %d: missing '=' separator\n", lineNo)
			continue
		}
		target := strings.TrimSpace(line[:idx])
		source := strings.TrimSpace(line[idx+1:])
		if target == "" || source == "" {
			if !lenient {
				return nil, fmt.Errorf("line %d: empty target or source", lineNo)
			}
			fmt.Fprintf(warn, "skipping line %d: empty target or source\n", lineNo)
			continue
		}
		if err := validateTarget(target); err != nil {
			if !lenient {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			fmt.Fprintf(warn, "line %d: %v, keeping as written\n", lineNo, err)
		}
		if err := validateSource(source); err != nil {
			if !lenient {
				return nil, fmt.Errorf("line %d: %w", lineNo, err)
			}
			fmt.Fprintf(warn, "line %d: %v, keeping as written\n", lineNo, err)
		}
		entries = append(entries, Entry{Target: target, Source: source, Line: lineNo})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func validateTarget(target string) error {
	if strings.HasPrefix(target, "~/") || strings.HasPrefix(target, "/") {
		return nil
	}
	return fmt.Errorf("target %q is not absolute or ~-rooted", target)
}

func validateSource(source string) error {
	if strings.HasPrefix(source, "/") || strings.HasPrefix(source, "~") {
		return fmt.Errorf("source %q must be relative to the dotfiles repo", source)
	}
	for _, part := range strings.Split(source, "/") {
		if part == ".." {
			return fmt.Errorf("source %q escapes the repo with '..'", source)
		}
	}
	return nil
}

// WriteLinkfile writes entries back out in the linkfile format, in
// the order they were given.
func WriteLinkfile(w io.Writer, entries []Entry) error {
	for _, e := range entries {
		if _, err := fmt.Fprintf(w, "%s = %s\n", e.Target, e.Source); err != nil {
			return err
		}
	}
	return nil
}
