// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Tests of the whole collection: all layers are registered and complete,
// every tool follows the general version rule, every layer file has its
// release choice as constants, and no feature uses distribution-specific
// code.

package features

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/majikmate/devcontainer-core/pkg/layer"
)

// The layers of this repository and the folder of their package.
var packages = map[string]string{
	"aliases":       "aliases",
	"deno":          "deno",
	"git":           "git",
	"github-cli":    "githubcli",
	"go":            "golang",
	"node":          "node",
	"prettier":      "prettier",
	"pure-prompt":   "pureprompt",
	"html-validate": "htmlvalidate",
}

func TestAllFeaturesRegistered(t *testing.T) {
	for name, dir := range packages {
		l, err := layer.Get(name)
		if err != nil {
			t.Errorf("feature %s: %v", name, err)
			continue
		}
		if l.Summary == "" || l.Install == nil || l.Test == nil {
			t.Errorf("feature %s: summary, install and test are required", name)
		}
		if want := "github.com/majikmate/devcontainer-features/" + dir; l.Package != want {
			t.Errorf("feature %s: registered by %s, want %s", name, l.Package, want)
		}
		for _, tool := range l.Tools {
			if tool.Name == "" || tool.Arg == "" || tool.Source == nil {
				t.Errorf("feature %s: tool %+v is incomplete", name, tool)
			}
		}
	}
	if got := len(layer.All()); got != len(packages) {
		t.Errorf("%d layers registered, want %d", got, len(packages))
	}
}

// TestVersionsDeclared checks the general version rule for every tool: a
// complete source, a valid channel, and an end-of-life rule for a pinned
// line.
func TestVersionsDeclared(t *testing.T) {
	for name := range packages {
		l, err := layer.Get(name)
		if err != nil {
			t.Fatal(err)
		}
		for _, tool := range l.Tools {
			if tool.Source == nil || tool.Source.Releases == nil || tool.Source.Name == "" {
				t.Errorf("tool %s: no complete source", tool.Name)
				continue
			}
			if _, err := tool.Effective(tool.Version); err != nil {
				t.Errorf("tool %s: %v", tool.Name, err)
			}
			if tool.Version.Pin != "" && tool.Source.Support == nil {
				t.Errorf("tool %s: pinned without an end-of-life rule", tool.Name)
			}
		}
	}
}

// TestReleaseChoiceConstants checks that the file of every layer with tools
// declares its release choice as constants <tool>Pin and <tool>Channel, as
// the first declaration after the imports.
func TestReleaseChoiceConstants(t *testing.T) {
	for name, dir := range packages {
		l, _ := layer.Get(name)
		own := 0
		for _, tool := range l.Tools {
			if tool.Follows == "" {
				own++
			}
		}
		file := filepath.Join(dir, dir+".go")
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		var first *ast.GenDecl
		for _, decl := range f.Decls {
			if g, ok := decl.(*ast.GenDecl); ok && g.Tok == token.IMPORT {
				continue
			}
			first, _ = decl.(*ast.GenDecl)
			break
		}
		if own == 0 {
			continue
		}
		var pins, channels int
		if first != nil && first.Tok == token.CONST {
			for _, spec := range first.Specs {
				for _, id := range spec.(*ast.ValueSpec).Names {
					switch {
					case strings.HasSuffix(id.Name, "Pin"):
						pins++
					case strings.HasSuffix(id.Name, "Channel"):
						channels++
					}
				}
			}
		}
		if pins != own || channels != own {
			t.Errorf("%s: the first declaration after the imports must be the constants <tool>Pin and <tool>Channel of the %d tools with an own release choice (found %d pins, %d channels)", file, own, pins, channels)
		}
	}
}

// Features are distribution-independent: they must not use the Debian
// package of the framework (apt). What a feature needs from the system, it
// declares as a needed layer of devcontainer-core.
//
// The test checks the folders that the go command uses for packages: it
// skips folders whose name starts with "." or "_" and folders named
// "testdata", like the go command. So it does not check the checkout of the
// tooling (.devcon) in the release workflow.
func TestNoDistributionSpecificImports(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name := d.Name(); path != "." && (strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if strings.HasSuffix(p, "/pkg/debian") {
				t.Errorf("%s imports %s; features must be distribution-independent", path, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// constantLink matches a README link to the release choice constants of a
// layer file, with its line range, for example "](deno/deno.go#L36-L39)".
var constantLink = regexp.MustCompile(`\]\(([a-z]+/[a-z]+\.go)#L([0-9]+)-L([0-9]+)\)`)

// TestReadmeLinksToConstants: the README links to the pin and channel
// constants of every layer with a line range; the range must still be the
// const block with the constants. The READMEs of devcontainer-core and of the
// images use the same ranges.
func TestReadmeLinksToConstants(t *testing.T) {
	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatal(err)
	}
	links := constantLink.FindAllStringSubmatch(string(readme), -1)
	if len(links) == 0 {
		t.Fatal("the README has no link to the constants of a layer")
	}
	for _, m := range links {
		source, err := os.ReadFile(m[1])
		if err != nil {
			t.Errorf("README link %s: %v", m[0], err)
			continue
		}
		lines := strings.Split(string(source), "\n")
		from, _ := strconv.Atoi(m[2])
		to, _ := strconv.Atoi(m[3])
		if from < 1 || to > len(lines) || from >= to || lines[from-1] != "const (" || lines[to-1] != ")" ||
			!strings.Contains(strings.Join(lines[from-1:to], "\n"), "Pin ") {
			t.Errorf("README link %s does not point to the const block with the pin and channel constants", m[0])
		}
	}
}
