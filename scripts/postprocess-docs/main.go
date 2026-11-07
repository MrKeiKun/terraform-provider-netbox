// Command postprocess-docs finishes the docs written by tfplugindocs.
//
// The docs tool cannot set the registry subcategory of a page from the provider
// schema. The resource and data source descriptions therefore start with a
// marker like
//
//	:meta:subcategory:Extras:Tags are user-defined labels ...
//
// For every page that contains such a marker this command writes the name into
// the "subcategory" field of the page header and removes the marker from the
// text. Pages that already have a subcategory from a custom template keep it.
//
// Usage: go run ./scripts/postprocess-docs <docs directory>
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	markerRe      = regexp.MustCompile(`:meta:subcategory:([^:\n]+):`)
	subcategoryRe = regexp.MustCompile(`(?m)^subcategory: "([^"\n]*)"$`)
)

// process rewrites the content of one docs page. It reports whether the page
// was changed.
func process(content string) (string, bool) {
	match := markerRe.FindStringSubmatch(content)
	if match == nil {
		return content, false
	}
	name := strings.TrimSpace(match[1])

	// Only the page header at the top of the file has a subcategory field.
	header, body := splitHeader(content)
	if current := subcategoryRe.FindStringSubmatch(header); current != nil && current[1] == "" {
		header = subcategoryRe.ReplaceAllString(header, fmt.Sprintf("subcategory: %q", name))
	}

	result := markerRe.ReplaceAllString(header+body, "")
	return result, result != content
}

// splitHeader splits a page into its "---" delimited header and the rest.
func splitHeader(content string) (header, body string) {
	if !strings.HasPrefix(content, "---\n") {
		return "", content
	}
	end := strings.Index(content[4:], "\n---\n")
	if end < 0 {
		return "", content
	}
	cut := 4 + end + len("\n---\n")
	return content[:cut], content[cut:]
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: postprocess-docs <docs directory>")
		os.Exit(2)
	}

	changed := 0
	err := filepath.WalkDir(os.Args[1], func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out, ok := process(string(raw))
		if !ok {
			return nil
		}
		changed++
		return os.WriteFile(path, []byte(out), 0o644)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "postprocess-docs:", err)
		os.Exit(1)
	}
	fmt.Printf("postprocess-docs: updated %d pages\n", changed)
}
