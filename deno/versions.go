// This file lists the Deno releases for the general version rule of
// devcontainer-core: the release tags, the channels "lts" and "stable", and
// the end of life of a release line (a Deno major release). The release
// choice of the feature (pin and channel) is at the top of deno.go.

package deno

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/sys"
	"github.com/majikmate/devcontainer-core/pkg/versions"
)

const (
	denoReleases  = "https://github.com/denoland/deno/releases"
	denoLatestURL = "https://dl.deno.land/release-latest.txt"
	denoLTSFile   = "https://dl.deno.land/release-lts-latest.txt"
)

// denoSource lists the Deno releases (version tags up to the newest release
// on dl.deno.land). Channels
// (https://docs.deno.com/runtime/fundamentals/stability_and_releases/):
//
//   - "lts" is exactly the release that Deno names in release-lts-latest.txt
//     (the same file as "deno upgrade lts"): Deno promotes one release of the
//     LTS line to LTS and builds it again with the channel "long term
//     support". Newer patch releases of the same minor line (for example
//     v2.9.7 after the LTS release v2.9.3) are stable releases.
//   - "stable" are all releases.
//
// A release line is a major release (for example 2). Deno supports only its
// newest major release, so a line ends when Deno publishes a newer one.
var denoSource = &layer.Source{
	Name: "Deno releases (" + denoReleases + ", " + denoLTSFile + ")",
	Channels: []layer.Channel{
		{Name: "lts", Label: "long-term support, exactly the release in " + denoLTSFile},
		{Name: "stable", Label: "all releases"},
	},
	Policy: "a line is a Deno major release; it ends when Deno publishes a newer major release (" + denoReleases + ")",
	Releases: func() ([]layer.Release, error) {
		latest, err := readLine(denoLatestURL)
		if err != nil {
			return nil, err
		}
		tags, err := sys.Output("git", "ls-remote", "--tags", "--refs", "https://github.com/denoland/deno", "refs/tags/v*")
		if err != nil {
			return nil, err
		}
		// Without the LTS file, the channel lts has no release (the release
		// check reports it)
		lts, _ := readLine(denoLTSFile)
		return parseDenoReleases(tags, latest, lts), nil
	},
	Support: func(line string) error {
		if !majorLine.MatchString(line) {
			return fmt.Errorf("Deno line %s: a Deno line is a major version, for example 2", line)
		}
		latest, err := readLine(denoLatestURL)
		if err != nil {
			return err
		}
		return denoSupport(latest, line)
	},
}

// majorLine matches a Deno release line: a major version, for example 2.
var majorLine = regexp.MustCompile(`^[0-9]+$`)

// readLine reads a file with one line, for example "v2.9.3".
func readLine(url string) (string, error) {
	data, err := sys.Get(url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

var denoTag = regexp.MustCompile(`refs/tags/(v[0-9]+\.[0-9]+\.[0-9]+)$`)

// parseDenoReleases returns the release tags vX.Y.Z of "git ls-remote --tags"
// up to the newest release latest (a tag can exist before its files are on
// dl.deno.land), highest first. All releases are in the channel "stable",
// the release lts also in "lts".
func parseDenoReleases(tags, latest, lts string) []layer.Release {
	var releases []layer.Release
	for _, line := range strings.Split(tags, "\n") {
		m := denoTag.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil || versions.Compare(m[1], latest) > 0 {
			continue
		}
		channels := []string{"stable"}
		if m[1] == lts {
			channels = append(channels, "lts")
		}
		releases = append(releases, layer.Release{Version: m[1], Channels: channels})
	}
	sort.SliceStable(releases, func(i, j int) bool { return versions.Compare(releases[i].Version, releases[j].Version) > 0 })
	return releases
}

// denoSupport returns an *layer.EndOfLifeError when the newest Deno release
// has a higher major version than the line.
func denoSupport(latest, line string) error {
	major := strings.SplitN(strings.TrimPrefix(latest, "v"), ".", 2)[0]
	switch c := versions.Compare(line, major); {
	case c > 0:
		return fmt.Errorf("Deno %s is not released yet (newest: %s)", line, latest)
	case c == 0:
		return nil
	}
	return &layer.EndOfLifeError{
		Since:     "Deno " + latest + " was released; Deno supports only its newest major release",
		Source:    denoReleases,
		Supported: []string{major},
	}
}
