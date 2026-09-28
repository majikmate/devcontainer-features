// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Package deno is the layer deno: Deno, the JavaScript/TypeScript runtime
// and language server, from the release archive on dl.deno.land, and the
// VS Code extension for Deno.
//
// This file declares the layer and its installation; versions.go lists the
// Deno releases, their channels and the end of life of a release line.
// Importing the package registers the layer:
//
//	import _ "github.com/majikmate/devcontainer-features/deno"
package deno

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/majikmate/devcontainer-core/pkg/devcontainer"
	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/sys"
)

// Release choice of the tool deno: the pinned release line and the release
// channel (see layer.Config and versions.go). The feature installs the LTS
// release (the one of "deno upgrade lts") of the major release 2; a newer
// Deno major release ends the line and stops the release. The channel
// "stable" would install the newest release of the line instead. The
// devcontainer.json of an image can override both
// ("customizations": {"devcon": {"deno": {"pin": "...", "channel": "..."}}}).
const (
	denoPin     = "2"
	denoChannel = "lts"
)

// init registers the layer.
func init() {
	layer.Register(&layer.Layer{
		Name:    "deno",
		Summary: "Deno, the JavaScript/TypeScript runtime and language server",
		Needs:   []string{"user"},
		Tools: []layer.Tool{
			{Name: "deno", Arg: "DENO_VERSION", Source: denoSource, Version: layer.Config{Pin: denoPin, Channel: denoChannel}},
		},
		Metadata: devcontainer.Entry{
			// No upgrade message of the Deno CLI and of the language server in
			// VS Code: the feature decides the Deno version (a new image
			// brings the next release of the channel).
			"containerEnv": map[string]any{"DENO_NO_UPDATE_CHECK": "1"},
			"customizations": devcontainer.VSCode([]string{"denoland.vscode-deno"}, map[string]any{
				"deno.enable":            true,
				"deno.codeLens.test":     true,
				"deno.codeLens.testArgs": []string{"--allow-all", "--check=all"},
				"deno.testing.args":      []string{"--allow-all", "--check=all"},
			}),
		},
		Install: install,
		// Deno runs a TypeScript file with type check; the test records the
		// version and the channel of the build
		Test: func(t *layer.T) {
			dir, cleanup, err := sys.TempDir()
			if err == nil {
				defer cleanup()
				file := filepath.Join(dir, "check.ts")
				_ = os.WriteFile(file, []byte("const n: number = 1 + 1;\nconsole.log(n);\n"), 0o644)
				t.Check("run TypeScript with type check", t.Output("deno run", "deno", "run", "--check", file) == "2")
			}
			version := t.Output("deno version", "deno", "--version")
			t.Version("deno", "v"+layer.FindVersion(version, `deno ([0-9.]+)`))
			// The channel of the build, for example "long term support" or "stable"
			t.Version("deno-channel", layer.FindVersion(version, `deno [0-9.]+ \(([^,]+),`))
		},
	})
}

// install downloads the release archive of the chosen version from
// dl.deno.land, checks it against the published SHA-256 checksum and
// installs the binary /usr/local/bin/deno.
func install(e *layer.Env) error {
	version, err := e.Version("deno")
	if err != nil {
		return err
	}
	version = "v" + strings.TrimPrefix(version, "v")
	target := map[string]string{"amd64": "x86_64-unknown-linux-gnu", "arm64": "aarch64-unknown-linux-gnu"}[runtime.GOARCH]
	if target == "" {
		return fmt.Errorf("deno: unsupported architecture %s", runtime.GOARCH)
	}
	dir, cleanup, err := sys.TempDir()
	if err != nil {
		return err
	}
	defer cleanup()
	name := "deno-" + target + ".zip"
	url := "https://dl.deno.land/release/" + version + "/" + name
	sys.Logf("Installing Deno %s", version)
	archive := filepath.Join(dir, name)
	if err := sys.Download(url, archive); err != nil {
		return err
	}
	sum, err := sys.Get(url + ".sha256sum")
	if err != nil {
		return err
	}
	if err := sys.VerifySHA256(archive, strings.ToLower(regexp.MustCompile(`[0-9a-fA-F]{64}`).FindString(string(sum)))); err != nil {
		return err
	}
	if err := sys.ExtractZip(archive, dir); err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(dir, "deno"))
	if err != nil {
		return err
	}
	return sys.WriteFile("/usr/local/bin/deno", string(data), 0o755)
}
