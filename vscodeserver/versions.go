// SPDX-License-Identifier: MIT
// © 2026 Hannes Stauss (scalarion@nimblescape.com)
// Licensed under the MIT License. See LICENSE in the repository root for details.

// This file lists the VS Code releases for the general version rule of
// devcontainer-core and reads the download data of a release (archive URL,
// SHA-256 checksum and commit) from the update service of VS Code. The
// release choice of the feature (pin and channel) is at the top of
// vscodeserver.go.

package vscodeserver

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/sys"
)

// updateService is the update service of VS Code.
const updateService = "https://update.code.visualstudio.com"

// vscodeSource lists the stable VS Code releases, newest first. A release
// line is a VS Code minor release (for example 1.105); Microsoft supports
// only the newest release, so the image follows it (empty pin).
var vscodeSource = &layer.Source{
	Name:   "VS Code releases (" + updateService + "/api/releases/stable)",
	Policy: "only the newest VS Code release gets updates (https://code.visualstudio.com/docs/supporting/faq)",
	Releases: func() ([]layer.Release, error) {
		data, err := sys.Get(updateService + "/api/releases/stable")
		if err != nil {
			return nil, err
		}
		return parseReleases(data)
	},
}

// productVersion matches a VS Code release, for example 1.105.1.
var productVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// parseReleases reads the answer of /api/releases/stable: a JSON list of
// product versions, newest first, for example ["1.105.1", "1.105.0"].
func parseReleases(data []byte) ([]layer.Release, error) {
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("VS Code releases: %w", err)
	}
	var releases []layer.Release
	for _, version := range list {
		if productVersion.MatchString(version) {
			releases = append(releases, layer.Release{Version: version})
		}
	}
	if len(releases) == 0 {
		return nil, fmt.Errorf("VS Code releases: the list has no release")
	}
	return releases, nil
}

// build is the download data of the VS Code Server of one release and
// platform.
type build struct {
	URL            string `json:"url"`            // the server archive (.tar.gz)
	SHA256         string `json:"sha256hash"`     // the SHA-256 checksum of the archive
	Commit         string `json:"version"`        // the commit of the release
	ProductVersion string `json:"productVersion"` // for example 1.105.1
}

// commit matches the commit of a VS Code release (40 hexadecimal digits).
var commit = regexp.MustCompile(`^[0-9a-f]{40}$`)

// parseBuild reads the answer of the update service for one release and
// platform, and checks that it is complete and belongs to the version.
func parseBuild(data []byte, version string) (build, error) {
	var b build
	if err := json.Unmarshal(data, &b); err != nil {
		return b, fmt.Errorf("VS Code Server %s: %w", version, err)
	}
	switch {
	case b.ProductVersion != version:
		return b, fmt.Errorf("VS Code Server %s: the update service answers with release %q", version, b.ProductVersion)
	case b.URL == "" || b.SHA256 == "":
		return b, fmt.Errorf("VS Code Server %s: the update service gives no archive URL or checksum", version)
	case !commit.MatchString(b.Commit):
		return b, fmt.Errorf("VS Code Server %s: the update service gives no commit (%q)", version, b.Commit)
	}
	return b, nil
}

// releaseBuild returns the download data of the VS Code Server of the
// release version for the platform (for example server-linux-x64). It asks
// for the release itself; when this fails, the data of the newest release is
// used if it is the same release.
func releaseBuild(version, platform string) (build, error) {
	data, err := sys.Get(updateService + "/api/versions/" + version + "/" + platform + "/stable")
	if err == nil {
		return parseBuild(data, version)
	}
	latest, latestErr := sys.Get(updateService + "/api/update/" + platform + "/stable/latest")
	if latestErr != nil {
		return build{}, err
	}
	b, latestErr := parseBuild(latest, version)
	if latestErr != nil {
		return build{}, fmt.Errorf("%w (newest release: %v)", err, latestErr)
	}
	return b, nil
}
