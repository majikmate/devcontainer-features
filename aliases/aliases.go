// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Package aliases is the layer aliases: shell aliases for all users in bash
// and zsh (ls, ll, grep, vs).
//
// The layer installs no tool, so it has no version, no pin and no channel.
// Importing the package registers the layer:
//
//	import _ "github.com/majikmate/devcontainer-features/aliases"
package aliases

import (
	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/shellrc"
)

// Shell aliases of all users (bash and zsh).
const aliases = `alias ls='ls -h --color=auto'
alias ll='ls -lah --color=auto'
alias grep='grep --color=auto'
alias vs='code -r .'`

// init registers the layer.
func init() {
	layer.Register(&layer.Layer{
		Name:    "aliases",
		Summary: "shell aliases: ls, ll, grep, vs",
		Needs:   []string{"user"},
		// The aliases are in the shared shell configuration of bash and zsh
		Install: func(e *layer.Env) error {
			return shellrc.Shared("aliases", aliases)
		},
		// Every alias exists in an interactive zsh and bash
		Test: func(t *layer.T) {
			for _, name := range []string{"ls", "ll", "grep", "vs"} {
				t.Command("zsh alias "+name, "zsh", "-ic", "alias "+name)
				t.Command("bash alias "+name, "bash", "-ic", "alias "+name)
			}
		},
	})
}
