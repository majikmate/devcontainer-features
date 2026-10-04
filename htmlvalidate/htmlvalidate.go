// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Package htmlvalidate is the layer html-validate: the VS Code extension
// HTML-validate, an offline HTML5 validator, and a global configuration
// /.htmlvalidate.json that fits the output of Prettier.
//
// The extension brings its own html-validate library and runs it in the
// container (pure JavaScript: no Java, every processor type). The layer
// installs no tool, so it has no version, no pin and no channel. Importing
// the package registers the layer:
//
//	import _ "github.com/majikmate/devcontainer-features/htmlvalidate"
package htmlvalidate

import (
	"encoding/json"
	"os"

	"github.com/majikmate/devcontainer-core/pkg/devcontainer"
	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/sys"
)

// configFile is the global configuration of html-validate. html-validate
// searches for a configuration from the folder of the checked file up to
// "/" and merges all files it finds, so every project uses this file and can
// still change single rules in its own .htmlvalidate.json.
const configFile = "/.htmlvalidate.json"

// config is the content of configFile: the recommended rules, without the
// style rules that conflict with the output of Prettier ("html-validate:prettier"
// switches off doctype-style, void-style and attr-quotes: Prettier writes
// <!doctype html> and <br />, both valid HTML).
var config = map[string]any{
	"extends": []string{"html-validate:recommended", "html-validate:prettier"},
}

// init registers the layer.
func init() {
	layer.Register(&layer.Layer{
		Name:    "html-validate",
		Summary: "HTML-validate, an offline HTML5 validator for VS Code, with a configuration that fits Prettier",
		Needs:   []string{"os"},
		Metadata: devcontainer.Entry{
			"customizations": devcontainer.VSCode([]string{"html-validate.vscode-html-validate"}, nil),
		},
		Install: func(e *layer.Env) error {
			return sys.WriteFile(configFile, content(), 0o644)
		},
		// The global configuration exists and has the expected content
		Test: func(t *layer.T) {
			data, err := os.ReadFile(configFile)
			t.Check("global configuration "+configFile, err == nil && string(data) == content())
		},
	})
}

// content returns the text of configFile.
func content() string {
	data, _ := json.MarshalIndent(config, "", "  ")
	return string(data) + "\n"
}
