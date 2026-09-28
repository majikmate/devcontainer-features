// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// This file has the version rule of the Go tools that follow the installed
// Go version (gopls, dlv, staticcheck, govulncheck). The Go release lines and
// their end of life are general (versions.GoReleases of devcontainer-core):
// a line is a Go major release 1.N; it ends when Go 1.(N+2) is released. The
// release choice of the feature (pin and channel) is at the top of golang.go.

package golang

import (
	"regexp"
	"strings"

	"github.com/majikmate/devcontainer-core/pkg/sys"
	"github.com/majikmate/devcontainer-core/pkg/versions"
)

// goModuleWorks reports whether a release of a Go module works with the
// installed Go version: its go.mod needs at most that version. The Go tools
// are built with the installed Go (GOTOOLCHAIN=local), so they must accept it.
func goModuleWorks(module string) func(release, goVersion string) (bool, error) {
	return func(release, goVersion string) (bool, error) {
		mod, err := sys.Get("https://proxy.golang.org/" + strings.ToLower(module) + "/@v/" + release + ".mod")
		if err != nil {
			return false, err
		}
		return versions.Compare(goDirective(string(mod)), goVersion) <= 0, nil
	}
}

// goDirective returns the Go version of the go directive of a go.mod file
// ("" when it has none: every Go version accepts it).
func goDirective(mod string) string {
	m := regexp.MustCompile(`(?m)^go\s+([0-9][0-9.]*)`).FindStringSubmatch(mod)
	if m == nil {
		return ""
	}
	return m[1]
}
