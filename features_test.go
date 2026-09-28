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
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/majikmate/devcontainer-core/pkg/layer"
)

// The layers of this repository and the folder of their package.
var packages = map[string]string{
	"aliases":     "aliases",
	"deno":        "deno",
	"git":         "git",
	"github-cli":  "githubcli",
	"go":          "golang",
	"node":        "node",
	"prettier":    "prettier",
	"pure-prompt": "pureprompt",
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
