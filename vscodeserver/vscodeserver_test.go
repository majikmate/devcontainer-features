// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// Tests of the folder handling of Sync (vscodeserver.go) in a temporary
// folder, without downloads.

package vscodeserver

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// names returns the sorted entry names of a folder.
func names(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var list []string
	for _, e := range entries {
		list = append(list, e.Name())
	}
	sort.Strings(list)
	return list
}

// TestLinkAndPrune: the link points to the server relative to the folder;
// prune deletes the servers and links of commits that are not wanted.
func TestLinkAndPrune(t *testing.T) {
	dir := t.TempDir()
	keepCommit, oldCommit := strings.Repeat("a", 40), strings.Repeat("b", 40)
	for _, c := range []string{keepCommit, oldCommit} {
		if err := os.MkdirAll(filepath.Join(dir, "bin", c, "bin"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "bin", c, "bin", "code-server"), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := link(dir, c); err != nil {
			t.Fatal(err)
		}
	}
	// A second link call keeps the link
	if err := link(dir, keepCommit); err != nil {
		t.Fatal(err)
	}
	server := filepath.Join(dir, "cli", "servers", "Stable-"+keepCommit, "server", "bin", "code-server")
	if _, err := os.Stat(server); err != nil {
		t.Errorf("server through the link: %v", err)
	}
	if target, _ := os.Readlink(filepath.Join(dir, "cli", "servers", "Stable-"+keepCommit, "server")); filepath.IsAbs(target) {
		t.Errorf("link target %q is absolute", target)
	}

	if err := prune(dir, map[string]bool{keepCommit: true}); err != nil {
		t.Fatal(err)
	}
	if got := names(t, filepath.Join(dir, "bin")); len(got) != 1 || got[0] != keepCommit {
		t.Errorf("servers after prune = %v", got)
	}
	if got := names(t, filepath.Join(dir, "cli", "servers")); len(got) != 1 || got[0] != "Stable-"+keepCommit {
		t.Errorf("links after prune = %v", got)
	}
}

// TestRemoveTemp: the temporary folders of an interrupted Sync are deleted,
// nothing else.
func TestRemoveTemp(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{tempPrefix + "123", "bin"} {
		if err := os.MkdirAll(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := removeTemp(dir); err != nil {
		t.Fatal(err)
	}
	if got := names(t, dir); len(got) != 1 || got[0] != "bin" {
		t.Errorf("entries = %v, want [bin]", got)
	}
}

// TestSyncKeep: Sync needs at least one release.
func TestSyncKeep(t *testing.T) {
	if err := Sync(t.TempDir(), 0); err == nil {
		t.Error("keep 0: no error")
	}
}
