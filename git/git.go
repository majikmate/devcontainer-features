// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Package git is the layer git: system-wide git settings for simple
// workflows. git itself comes with the layer os of devcontainer-core.
//
// The layer installs no tool, so it has no version, no pin and no channel.
// Importing the package registers the layer:
//
//	import _ "github.com/majikmate/devcontainer-features/git"
package git

import (
	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/sys"
)

// System-wide git settings for simple workflows: pull rebases local commits,
// and local changes are stashed during a rebase.
var gitSettings = [][2]string{
	{"pull.rebase", "true"},
	{"rebase.autoStash", "true"},
}

// init registers the layer.
func init() {
	layer.Register(&layer.Layer{
		Name:    "git",
		Summary: "system-wide git settings (rebase on pull, auto stash)",
		Needs:   []string{"os"},
		// The settings go to the system configuration (/etc/gitconfig)
		Install: func(e *layer.Env) error {
			for _, s := range gitSettings {
				if err := sys.Run(nil, "git", "config", "--system", s[0], s[1]); err != nil {
					return err
				}
			}
			return nil
		},
		// Every setting has its value; the test records the git version
		Test: func(t *layer.T) {
			for _, s := range gitSettings {
				value, _ := sys.Output("git", "config", "--system", s[0])
				t.Check(s[0]+" = "+s[1], value == s[1])
			}
			v, _ := sys.Output("git", "--version")
			t.Version("git", layer.LastWord(v))
		},
	})
}
