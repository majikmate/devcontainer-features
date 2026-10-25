// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Package ts is the layer ts: the command ts of the npm package
// @nimblescape/typescript (https://github.com/nimblescape/nimblescape-typescript),
// installed with npm in /usr/local, its Tab completion for zsh and bash, and
// the VS Code extension nimblescape TypeScript.
//
// ts checks a TypeScript file with tsc, compiles it and runs it with Node.js,
// with the global functions prompt(), alert() and confirm(). The layer
// installs no TypeScript: in the terminals of VS Code, the extension gives ts
// the TypeScript of VS Code; in another terminal, ts offers to install one.
//
// Importing the package registers the layer:
//
//	import _ "github.com/majikmate/devcontainer-features/ts"
package ts

import (
	"strings"

	"github.com/majikmate/devcontainer-core/pkg/devcontainer"
	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/sys"
	"github.com/majikmate/devcontainer-core/pkg/versions"
)

// Release choice of the tool nimblescape-typescript (the npm package
// @nimblescape/typescript): the pinned release line and the release channel
// (see layer.Config). An empty pin means no pin: the newest release. An empty
// channel means the default channel of the source (the npm dist-tag latest).
// The devcontainer.json of an image can override both
// ("customizations": {"devcon": {"nimblescape-typescript": {"pin": "...", "channel": "..."}}}).
const (
	tsPin     = ""
	tsChannel = ""
)

// Paths of ts: the npm prefix and the completion scripts. zsh finds _ts in
// its folder site-functions, which is in its default fpath; bash-completion
// loads ts when the command line needs it, from the folder
// bash-completion/completions of a folder in XDG_DATA_DIRS (default
// /usr/local/share:/usr/share).
const (
	tsPrefix       = "/usr/local"
	tsCommand      = tsPrefix + "/bin/ts"
	zshCompletion  = tsPrefix + "/share/zsh/site-functions/_ts"
	bashCompletion = tsPrefix + "/share/bash-completion/completions/ts"
)

// init registers the layer.
func init() {
	layer.Register(&layer.Layer{
		Name:    "ts",
		Summary: "the command ts (checks, compiles and runs TypeScript with tsc and Node.js) with Tab completion, and the VS Code extension nimblescape TypeScript",
		Needs:   []string{"node"},
		Tools: []layer.Tool{
			{Name: "nimblescape-typescript", Arg: "NIMBLESCAPE_TYPESCRIPT_VERSION", Source: versions.NPMPackage("@nimblescape/typescript"), Version: layer.Config{Pin: tsPin, Channel: tsChannel}},
		},
		Metadata: devcontainer.Entry{
			// The extension gives ts the TypeScript of VS Code, runs F5 with
			// the debug type ts and gives the editor the defaults of ts. A
			// global ts comes first on the PATH, so this layer decides the
			// version of ts in the terminals and in the debug sessions.
			"customizations": devcontainer.VSCode([]string{"nimblescape.nimblescape-typescript"}, nil),
		},
		Install: install,
		// ts runs, and the interactive shells complete its options with Tab
		Test: func(t *layer.T) {
			t.Version("nimblescape-typescript", strings.TrimPrefix(firstLine(t.Output("ts version", "ts", "--version")), "ts "))
			t.Command("ts prints its help", "ts", "--help")
			t.Command("Tab completion of ts in zsh", "zsh", "-ic", `(( ${+_comps[ts]} ))`)
			// bash-completion 2.12 and newer: _comp_load; older: _completion_loader
			t.Command("Tab completion of ts in bash", "bash", "-ic", `{ _comp_load ts || _completion_loader ts; } 2>/dev/null; complete -p ts`)
		},
	})
}

// install installs the npm package and writes the completion scripts that
// ts prints.
func install(e *layer.Env) error {
	version, err := e.Version("nimblescape-typescript")
	if err != nil {
		return err
	}
	// In /usr/local, not in the nvm folder, so ts does not depend on the
	// Node.js version that is selected with nvm.
	if err := sys.Run(nil, "npm", "install", "--global", "--prefix", tsPrefix,
		"--no-fund", "--no-audit", "--no-update-notifier",
		"@nimblescape/typescript@"+version); err != nil {
		return err
	}
	_ = sys.Run(nil, "npm", "cache", "clean", "--force")
	for shell, file := range map[string]string{"zsh": zshCompletion, "bash": bashCompletion} {
		script, err := sys.Output(tsCommand, "--completion", shell)
		if err != nil {
			return err
		}
		if err := sys.WriteFile(file, script+"\n", 0o644); err != nil {
			return err
		}
	}
	return nil
}

// firstLine returns the first line of a text.
func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")
	return strings.TrimSpace(line)
}
