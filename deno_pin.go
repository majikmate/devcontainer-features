package features

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/sys"
	"github.com/majikmate/devcontainer-core/pkg/versions"
)

// Deno release lines: a line is a major release (for example 2). Deno
// supports only its newest major release, so a line ends when Deno publishes
// a newer major release.
var denoPin = &layer.Pin{
	Arg:     "DENO_PIN",
	Example: "2",
	Policy:  "a line is a Deno major release; it ends when Deno publishes a newer major release (" + denoReleases + ")",
	Newest:  denoNewestIn,
	Support: func(line string) error {
		if !majorLine.MatchString(line) {
			return fmt.Errorf("DENO_PIN=%s: a Deno line is a major version, for example 2", line)
		}
		latest, err := denoLatest()
		if err != nil {
			return err
		}
		return denoSupport(latest, line)
	},
}

const denoReleases = "https://github.com/denoland/deno/releases"

// denoLatest returns the newest Deno release, for example "v2.5.6".
func denoLatest() (string, error) {
	data, err := sys.Get("https://dl.deno.land/release-latest.txt")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// denoNewestIn returns the newest release of a major line: the LTS release
// or the newest release when it is in the line (the same choice as without
// a pin), otherwise the highest version tag of the line in the Deno
// repository (an older major line).
func denoNewestIn(line string) (string, error) {
	if !majorLine.MatchString(line) {
		return "", fmt.Errorf("DENO_PIN=%s: a Deno line is a major version, for example 2", line)
	}
	if v, err := denoLTS(); err == nil && layer.InLine(v, line) {
		return v, nil
	}
	if v, err := denoLatest(); err == nil && layer.InLine(v, line) {
		return v, nil
	}
	out, err := sys.Output("git", "ls-remote", "--tags", "--refs", "https://github.com/denoland/deno", "refs/tags/v"+line+".*")
	if err != nil {
		return "", err
	}
	return highestTag(out, line)
}

// highestTag returns the highest tag vX.Y.Z of a line in the output of "git
// ls-remote --tags".
// denoLTS returns the newest release of the Deno LTS line. Deno's LTS
// channel is a minor line (for example 2.9) that gets backported patch
// releases; the file release-lts-latest.txt names the release with which the
// channel started (for example v2.9.3), not the newest patch release (for
// example v2.9.7).
func denoLTS() (string, error) {
	start, err := versions.DenoLTS()
	if err != nil {
		return "", err
	}
	out, err := sys.Output("git", "ls-remote", "--tags", "--refs", "https://github.com/denoland/deno", "refs/tags/"+ltsLine(start)+".*")
	if err != nil {
		return start, nil
	}
	return newestLTS(start, out), nil
}

// ltsLine returns the minor line of a release as a tag prefix, for example
// "v2.9" for v2.9.3.
func ltsLine(release string) string {
	parts := strings.SplitN(strings.TrimPrefix(release, "v"), ".", 3)
	if len(parts) < 2 {
		return "v" + parts[0]
	}
	return "v" + parts[0] + "." + parts[1]
}

// newestLTS returns the newest release in the tag list (git ls-remote) that
// belongs to the minor line of start, or start when there is none.
func newestLTS(start, out string) string {
	best, err := highestTag(out, strings.TrimPrefix(ltsLine(start), "v"))
	if err != nil || compareVersions(best, start) < 0 {
		return start
	}
	return best
}

func highestTag(out, line string) (string, error) {
	tag := regexp.MustCompile(`refs/tags/(v[0-9]+\.[0-9]+\.[0-9]+)$`)
	best := ""
	for _, l := range strings.Split(out, "\n") {
		m := tag.FindStringSubmatch(strings.TrimSpace(l))
		if m != nil && layer.InLine(m[1], line) && (best == "" || compareVersions(m[1], best) > 0) {
			best = m[1]
		}
	}
	if best == "" {
		return "", fmt.Errorf("Deno %s has no release", line)
	}
	return best, nil
}

// denoSupport returns an *layer.EndOfLifeError when the newest Deno release
// has a higher major version than the line.
func denoSupport(latest, line string) error {
	major := strings.SplitN(strings.TrimPrefix(latest, "v"), ".", 2)[0]
	switch c := compareVersions(line, major); {
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
