package features

import (
	"os"
	"path/filepath"

	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/shellrc"
	"github.com/majikmate/devcontainer-core/pkg/sys"
	"github.com/majikmate/devcontainer-core/pkg/versions"
)

const pureDir = "/usr/local/share/zsh/pure"

func init() {
	layer.Register(&layer.Layer{
		Name:    "pure-prompt",
		Summary: "Pure prompt for zsh (https://github.com/sindresorhus/pure)",
		Needs:   []string{"user"},
		Tools: []layer.Tool{
			{Name: "pure", Arg: "PURE_VERSION", Newest: func() (string, error) { return versions.GitHubRelease("sindresorhus/pure") }},
		},
		Install: func(e *layer.Env) error {
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
			// Pure shows user@host only over SSH, in a container outside of
			// GitHub Codespaces, or as root (flag psvar[13], set once by
			// "prompt pure"). The flag after "prompt pure" shows it always,
			// also in Codespaces.
			return shellrc.Zsh("prompt", `fpath+=("`+pureDir+`")
autoload -Uz promptinit
promptinit
zstyle :prompt:pure:host show yes
zstyle :prompt:pure:git:stash show yes
prompt pure
psvar[13]=1`)
		},
		Test: func(t *layer.T) {
			t.Command("Pure prompt active in zsh", "zsh", "-ic", `[[ "$prompt_theme[1]" == pure ]]`)
			t.Command("Pure prompt shows user@host, also in Codespaces", "env", "CODESPACES=true", "zsh", "-ic", `[[ "$psvar[13]" == 1 ]]`)
			data, _ := os.ReadFile(filepath.Join(pureDir, "VERSION"))
			t.Version("pure", string(data))
		},
	})
}
