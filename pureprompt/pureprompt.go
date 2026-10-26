// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Package pureprompt is the layer pure-prompt: the Pure prompt for zsh
// (https://github.com/sindresorhus/pure), from the files of a release tag.
//
// Importing the package registers the layer:
//
//	import _ "github.com/majikmate/devcontainer-features/pureprompt"
package pureprompt

import (
	"os"
	"path/filepath"

	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/shellrc"
	"github.com/majikmate/devcontainer-core/pkg/sys"
	"github.com/majikmate/devcontainer-core/pkg/versions"
)

// Release choice of the tool pure: the pinned release line and the release
// channel (see layer.Config). An empty pin means no pin: the newest release.
// An empty channel means the default channel of the source (GitHub releases
// have no channels). The devcontainer.json of an image can override both
// ("customizations": {"devcon": {"pure": {"pin": "...", "channel": "..."}}}).
const (
	purePin     = ""
	pureChannel = ""
)

// pureDir is the folder of the prompt files; zsh finds them through fpath.
const pureDir = "/usr/local/share/zsh/pure"

// init registers the layer.
func init() {
	layer.Register(&layer.Layer{
		Name:    "pure-prompt",
		Summary: "Pure prompt for zsh (https://github.com/sindresorhus/pure)",
		Needs:   []string{"user"},
		Tools: []layer.Tool{
			{Name: "pure", Arg: "PURE_VERSION", Source: versions.GitHubRepository("sindresorhus/pure"), Version: layer.Config{Pin: purePin, Channel: pureChannel}},
		},
		Install: install,
		// Pure is the active prompt of an interactive zsh. The tests run in a
		// container as the user of the image (not root, no SSH): the prompt
		// shows no user@host, and CODESPACES stays unset after the setup.
		Test: func(t *layer.T) {
			t.Command("Pure prompt active in zsh", "zsh", "-ic", `[[ "$prompt_theme[1]" == pure ]]`)
			t.Command("no user@host in the prompt in a container", "env", "-u", "CODESPACES", "-u", "SSH_CONNECTION", "zsh", "-ic", `[[ "$(print -P -- "$PROMPT")" != *"@${HOST%%.*}"* ]]`)
			t.Command("CODESPACES unset after the prompt setup", "env", "-u", "CODESPACES", "zsh", "-ic", `[[ -z "${CODESPACES+set}" ]]`)
			data, _ := os.ReadFile(filepath.Join(pureDir, "VERSION"))
			t.Version("pure", string(data))
		},
	})
}

// install downloads the prompt files of the release tag and activates the
// prompt in the zsh configuration.
func install(e *layer.Env) error {
	version, err := e.Version("pure")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(pureDir, 0o755); err != nil {
		return err
	}
	// zsh loads a prompt from the file prompt_<name>_setup in fpath
	files := map[string]string{"pure.zsh": "prompt_pure_setup", "async.zsh": "async"}
	for source, target := range files {
		url := "https://raw.githubusercontent.com/sindresorhus/pure/" + version + "/" + source
		if err := sys.Download(url, filepath.Join(pureDir, target)); err != nil {
			return err
		}
		if err := os.Chmod(filepath.Join(pureDir, target), 0o644); err != nil {
			return err
		}
	}
	if err := sys.WriteFile(filepath.Join(pureDir, "VERSION"), version+"\n", 0o644); err != nil {
		return err
	}
	// Pure shows user@host over SSH, as root, and in a container, but not in
	// GitHub Codespaces (CODESPACES is set there). The image always runs as a
	// container, in a Codespace or elsewhere (for example a devspace). With
	// CODESPACES set for the setup, the prompt is the same everywhere:
	// user@host only over SSH or as root. Pure reads CODESPACES only in its
	// setup, and zsh sets the variable only during the call of prompt.
	return shellrc.Zsh("prompt", `fpath+=("`+pureDir+`")
autoload -Uz promptinit
promptinit
zstyle :prompt:pure:host show yes
zstyle :prompt:pure:git:stash show yes
CODESPACES=${CODESPACES:-true} prompt pure`)
}
