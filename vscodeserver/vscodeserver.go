// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Package vscodeserver keeps the VS Code Servers of the newest VS Code
// releases in a folder outside the images. It is not a layer: no image
// contains a VS Code Server.
//
// The monitor on a remote VM runs Sync (command "devcon vscode-server sync")
// on a schedule and keeps the servers in a Docker volume; the remote build
// mounts this volume into the containers. The Dev Containers extension of a
// VS Code with one of these releases then finds its server and starts it
// without a download. Any other release is downloaded by the extension as
// usual.
//
// Layout of the folder (the same as in ~/.vscode-server, so the folder can be
// linked or mounted there):
//
//	bin/<commit>/                        the server of one release
//	cli/servers/Stable-<commit>/server   link to bin/<commit> (newer layout)
//
// This file holds Sync and the folder handling; versions.go lists the VS Code
// releases and reads the download data of a release.
package vscodeserver

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/majikmate/devcontainer-core/pkg/sys"
)

// DefaultKeep is the number of VS Code releases whose servers Sync keeps:
// the newest release of each of the three newest minor versions (for
// example 1.139.1, 1.138.2 and 1.137.0).
const DefaultKeep = 3

// tempPrefix starts the name of the temporary folders of Sync in the folder;
// a server becomes visible only by the rename of its complete folder.
const tempPrefix = ".sync-"

// platform returns the platform name of the VS Code Server downloads for the
// architecture of this machine.
func platform() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "server-linux-x64", nil
	case "arm64":
		return "server-linux-arm64", nil
	}
	return "", fmt.Errorf("vscode-server: unsupported architecture %s", runtime.GOARCH)
}

// Sync makes dir hold the servers of the newest release of each of the keep
// newest VS Code minor versions: it downloads the missing servers (with a
// check of the SHA-256 checksum of the update service) and deletes all other
// servers in dir.
func Sync(dir string, keep int) error {
	if keep < 1 {
		return fmt.Errorf("vscode-server: keep at least one release (keep=%d)", keep)
	}
	name, err := platform()
	if err != nil {
		return err
	}
	all, err := releases()
	if err != nil {
		return err
	}
	versions := newestPerMinor(all, keep)
	if err := os.MkdirAll(filepath.Join(dir, "bin"), 0o755); err != nil {
		return err
	}
	if err := removeTemp(dir); err != nil {
		return err
	}
	wanted := map[string]bool{}
	for _, version := range versions {
		build, err := releaseBuild(version, name)
		if err != nil {
			return err
		}
		wanted[build.Commit] = true
		if exists(filepath.Join(dir, "bin", build.Commit)) {
			sys.Logf("VS Code Server %s (commit %s): present", build.ProductVersion, build.Commit)
		} else if err := install(dir, build); err != nil {
			return err
		}
		if err := link(dir, build.Commit); err != nil {
			return err
		}
	}
	return prune(dir, wanted)
}

// install downloads the server archive of one release, checks it and moves
// the unpacked server to dir/bin/<commit>.
func install(dir string, b build) error {
	temp, err := os.MkdirTemp(dir, tempPrefix)
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)

	sys.Logf("VS Code Server %s (commit %s): download", b.ProductVersion, b.Commit)
	archive := filepath.Join(temp, "vscode-server.tar.gz")
	if err := sys.Download(b.URL, archive); err != nil {
		return err
	}
	if err := sys.VerifySHA256(archive, b.SHA256); err != nil {
		return err
	}
	unpacked := filepath.Join(temp, "unpacked")
	if err := sys.ExtractTarGz(archive, unpacked); err != nil {
		return err
	}
	// The archive has one top folder, for example vscode-server-linux-x64
	entries, err := os.ReadDir(unpacked)
	if err != nil {
		return err
	}
	if len(entries) != 1 || !entries[0].IsDir() {
		return fmt.Errorf("vscode-server %s: the archive must have one top folder", b.ProductVersion)
	}
	// The containers read the server as another user than the monitor
	if err := os.Chmod(filepath.Join(unpacked, entries[0].Name()), 0o755); err != nil {
		return err
	}
	return os.Rename(filepath.Join(unpacked, entries[0].Name()), filepath.Join(dir, "bin", b.Commit))
}

// link creates dir/cli/servers/Stable-<commit>/server, a relative link to
// dir/bin/<commit>, so the link also works where the folder is mounted.
func link(dir, commit string) error {
	path := filepath.Join(dir, "cli", "servers", "Stable-"+commit, "server")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	target := filepath.Join("..", "..", "..", "bin", commit)
	if current, err := os.Readlink(path); err == nil && current == target {
		return nil
	}
	_ = os.Remove(path)
	return os.Symlink(target, path)
}

// prune deletes the servers and links in dir whose commit is not wanted.
func prune(dir string, wanted map[string]bool) error {
	servers, err := os.ReadDir(filepath.Join(dir, "bin"))
	if err != nil {
		return err
	}
	for _, e := range servers {
		if !wanted[e.Name()] {
			sys.Logf("VS Code Server commit %s: delete", e.Name())
			if err := os.RemoveAll(filepath.Join(dir, "bin", e.Name())); err != nil {
				return err
			}
		}
	}
	links, err := os.ReadDir(filepath.Join(dir, "cli", "servers"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range links {
		if !wanted[strings.TrimPrefix(e.Name(), "Stable-")] {
			if err := os.RemoveAll(filepath.Join(dir, "cli", "servers", e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

// removeTemp deletes the temporary folders of an interrupted Sync.
func removeTemp(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), tempPrefix) {
			if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

// exists reports whether path exists.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
