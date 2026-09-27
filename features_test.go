package features

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/majikmate/devcontainer-core/pkg/layer"
)

// The features of this repository, in the order of the documentation.
var names = []string{"aliases", "deno", "git", "github-cli", "go", "node", "prettier", "pure-prompt"}

func TestAllFeaturesRegistered(t *testing.T) {
	for _, name := range names {
		l, err := layer.Get(name)
		if err != nil {
			t.Errorf("feature %s: %v", name, err)
			continue
		}
		if l.Summary == "" || l.Install == nil || l.Test == nil {
			t.Errorf("feature %s: summary, install and test are required", name)
		}
		for _, tool := range l.Tools {
			if tool.Name == "" || tool.Arg == "" || tool.Source == nil {
				t.Errorf("feature %s: tool %+v is incomplete", name, tool)
			}
		}
	}
	if got := len(layer.All()); got != len(names) {
		t.Errorf("%d layers registered, want %d", got, len(names))
	}
}

// Features are distribution-independent: they must not use the Debian
// package of the framework (apt). What a feature needs from the system, it
// declares as a needed layer of devcontainer-core.
func TestNoDistributionSpecificImports(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if strings.HasSuffix(path, "/pkg/debian") {
				t.Errorf("%s imports %s; features must be distribution-independent", file, path)
			}
		}
	}
}
