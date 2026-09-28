// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Package vscodeserver is the layer vscode-server: the VS Code Server of a
// VS Code release, installed in the image for the development user. The Dev
// Containers extension of a VS Code with the same release (commit) finds the
// server in the image and starts it without a download, so a new container
// is ready sooner.
//
// This file declares the layer and its installation; versions.go lists the
// VS Code releases and reads the download data of a release. Importing the
// package registers the layer:
//
//	import _ "github.com/majikmate/devcontainer-features/vscodeserver"
package vscodeserver

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/state"
	"github.com/majikmate/devcontainer-core/pkg/sys"
)

// Release choice of the tool vscode-server: the pinned release line and the
// release channel (see layer.Config). An empty pin means no pin: the newest
// VS Code release, which matches the VS Code of most users. An empty channel
// means the default channel of the source (the VS Code releases have no
// channels here: only stable releases). The devcontainer.json of an image can
// override both
// ("customizations": {"devcon": {"vscode-server": {"pin": "...", "channel": "..."}}}).
const (
	vscodeServerPin     = ""
	vscodeServerChannel = ""
)

// serverDir is the folder of the servers in the home folder of the user; the
// Dev Containers extension installs a server in serverDir/bin/<commit>.
const serverDir = ".vscode-server"

// init registers the layer.
func init() {
	layer.Register(&layer.Layer{
		Name:    "vscode-server",
		Summary: "VS Code Server of the newest VS Code release, ready for the Dev Containers extension",
		Needs:   []string{"user"},
		Tools: []layer.Tool{
			{Name: "vscode-server", Arg: "VSCODE_SERVER_VERSION", Source: vscodeSource, Version: layer.Config{Pin: vscodeServerPin, Channel: vscodeServerChannel}},
		},
		Install: install,
		Test:    test,
	})
}

// platform returns the platform name of the VS Code Server downloads for the
// architecture of the image.
func platform() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "server-linux-x64", nil
	case "arm64":
		return "server-linux-arm64", nil
	}
	return "", fmt.Errorf("vscode-server: unsupported architecture %s", runtime.GOARCH)
}

// install downloads the server archive of the chosen VS Code release, checks
// it against the SHA-256 checksum of the update service and unpacks it into
// ~/.vscode-server/bin/<commit> of the development user. The same server is
// also linked at ~/.vscode-server/cli/servers/Stable-<commit>/server, the
// folder of the newer server layout.
func install(e *layer.Env) error {
	version, err := e.Version("vscode-server")
	if err != nil {
		return err
	}
	name, err := platform()
	if err != nil {
		return err
	}
	build, err := releaseBuild(version, name)
	if err != nil {
		return err
	}
	u, err := sys.LookupUser(state.User())
	if err != nil {
		return err
	}
	dir, cleanup, err := sys.TempDir()
	if err != nil {
		return err
	}
	defer cleanup()

	sys.Logf("Installing VS Code Server %s (commit %s)", build.ProductVersion, build.Commit)
	archive := filepath.Join(dir, "vscode-server.tar.gz")
	if err := sys.Download(build.URL, archive); err != nil {
		return err
	}
	if err := sys.VerifySHA256(archive, build.SHA256); err != nil {
		return err
	}
	unpacked := filepath.Join(dir, "unpacked")
	if err := sys.ExtractTarGz(archive, unpacked); err != nil {
		return err
	}
	// The archive has one top folder, for example vscode-server-linux-x64
	entries, err := os.ReadDir(unpacked)
	if err != nil {
		return err
	}
	if len(entries) != 1 || !entries[0].IsDir() {
		return fmt.Errorf("vscode-server: the archive must have one top folder")
	}

	root := filepath.Join(u.Home, serverDir)
	target := filepath.Join(root, "bin", build.Commit)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	if err := os.Rename(filepath.Join(unpacked, entries[0].Name()), target); err != nil {
		return err
	}
	link := filepath.Join(root, "cli", "servers", "Stable-"+build.Commit, "server")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		return err
	}
	if err := os.Symlink(target, link); err != nil && !os.IsExist(err) {
		return err
	}
	return sys.ChownTree(root, u.UID, u.GID)
}

// test checks that the server of one release is installed for the
// development user, that it runs, and records its version.
func test(t *layer.T) {
	u, err := sys.LookupUser(state.User())
	if err != nil {
		t.Check("development user exists", false)
		return
	}
	servers, _ := filepath.Glob(filepath.Join(u.Home, serverDir, "bin", "*", "bin", "code-server"))
	t.Check("one VS Code Server in ~/"+serverDir+"/bin", len(servers) == 1)
	if len(servers) != 1 {
		return
	}
	info, err := os.Stat(servers[0])
	t.Check("the development user owns the server", err == nil && ownerUID(info) == u.UID)
	// "code-server --version" prints the version, the commit and the architecture
	out := t.Output("VS Code Server runs", servers[0], "--version")
	t.Version("vscode-server", strings.TrimSpace(strings.SplitN(out, "\n", 2)[0]))
}

// ownerUID returns the user ID of the owner of a file (-1 when unknown).
func ownerUID(info os.FileInfo) int {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return int(st.Uid)
	}
	return -1
}
