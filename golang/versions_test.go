// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Tests of the rule for the Go tools that follow Go (versions.go) and of the
// declaration of the layer go.

package golang

import (
	"testing"

	"github.com/majikmate/devcontainer-core/pkg/layer"
)

func TestGoDirective(t *testing.T) {
	mod := "module golang.org/x/tools/gopls\n\ngo 1.28.0\n\ntoolchain go1.28.2\n"
	if got := goDirective(mod); got != "1.28.0" {
		t.Errorf("go directive = %q", got)
	}
	if got := goDirective("module x\n"); got != "" {
		t.Errorf("no go directive = %q", got)
	}
}

// TestLayer checks that go and golangci-lint use the release choice of the
// constants and that the Go tools follow go.
func TestLayer(t *testing.T) {
	l, err := layer.Get("go")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]layer.Config{
		"go":            {Pin: goPin, Channel: goChannel},
		"golangci-lint": {Pin: golangciLintPin, Channel: golangciLintChannel},
	}
	followers := map[string]bool{}
	for _, g := range goTools {
		followers[g.name] = true
	}
	for _, tool := range l.Tools {
		if followers[tool.Name] {
			if tool.Follows != "go" || tool.Works == nil || tool.Version != (layer.Config{}) {
				t.Errorf("tool %s does not follow go", tool.Name)
			}
			continue
		}
		if tool.Version != want[tool.Name] {
			t.Errorf("tool %s: configuration %+v, want %+v", tool.Name, tool.Version, want[tool.Name])
		}
	}
	if len(l.Tools) != len(want)+len(goTools) {
		t.Errorf("%d tools, want %d", len(l.Tools), len(want)+len(goTools))
	}
}
