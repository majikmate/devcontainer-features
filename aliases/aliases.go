// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Package aliases is the layer aliases: shell aliases for all users in bash
// and zsh (ls, ll, grep, vs), and the shell function code, which opens a
// folder in the same window of VS Code.
//
// The layer installs no tool, so it has no version, no pin and no channel.
// Importing the package registers the layer:
//
//	import _ "github.com/majikmate/devcontainer-features/aliases"
package aliases

import (
	"os"
	"path/filepath"

	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/shellrc"
	"github.com/majikmate/devcontainer-core/pkg/sys"
)

// Shell aliases of all users (bash and zsh), and the function code.
//
// In a terminal of VS Code, the command code of VS Code opens a folder in a
// new window, unless the option --reuse-window (-r) is given. The setting
// window.openFoldersInNewWindow would change that, but it is a setting of the
// computer of the user, which an image cannot set. So the function code adds
// --reuse-window: `code .` restarts the same window in the folder (the
// courses teach `code .` for a project in a subfolder). With -n or
// --new-window, a new window opens as before.
const aliases = `alias ls='ls -h --color=auto'
alias ll='ls -lah --color=auto'
alias grep='grep --color=auto'
alias vs='code -r .'
code() {
  case " $* " in
    *" -n "* | *" --new-window "*) command code "$@" ;;
    *) command code --reuse-window "$@" ;;
  esac
}`

// init registers the layer.
func init() {
	layer.Register(&layer.Layer{
		Name:    "aliases",
		Summary: "shell aliases: ls, ll, grep, vs; the function code, which opens a folder in the same window of VS Code",
		Needs:   []string{"user"},
		// The aliases are in the shared shell configuration of bash and zsh
		Install: func(e *layer.Env) error {
			return shellrc.Shared("aliases", aliases)
		},
		// Every alias exists in an interactive zsh and bash. The function code
		// adds --reuse-window, but not to -n: a stand-in for the command code
		// of VS Code prints its arguments
		Test: func(t *layer.T) {
			for _, name := range []string{"ls", "ll", "grep", "vs"} {
				t.Command("zsh alias "+name, "zsh", "-ic", "alias "+name)
				t.Command("bash alias "+name, "bash", "-ic", "alias "+name)
			}
			dir, cleanup, err := sys.TempDir()
			if err != nil {
				t.Check("temporary folder for the test of code", false)
				return
			}
			defer cleanup()
			_ = os.WriteFile(filepath.Join(dir, "code"), []byte("#!/bin/sh\necho \"$*\"\n"), 0o755)
			for _, shell := range []string{"zsh", "bash"} {
				for _, c := range [][2]string{{"code .", "--reuse-window ."}, {"vs", "--reuse-window -r ."}, {"code -n .", "-n ."}} {
					out := t.Output(shell+" "+c[0], shell, "-ic", "PATH="+dir+":$PATH; "+c[0])
					t.Check(shell+" "+c[0]+" runs code "+c[1], out == c[1])
				}
			}
		},
	})
}
