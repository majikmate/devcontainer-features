package features

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/majikmate/devcontainer-core/pkg/layer"
	"github.com/majikmate/devcontainer-core/pkg/sys"
)

// Go release lines: a line is a Go major release 1.N (for example 1.27). Go
// supports the two newest major releases, so the line 1.N ends when Go
// 1.(N+2) is released.
var goPin = &layer.Pin{
	Arg:     "GO_PIN",
	Example: "1.27",
	Policy:  "a line is a Go major release 1.N; it ends when Go 1.(N+2) is released (Go supports the two newest major releases, " + goPolicy + ")",
	Newest: func(line string) (string, error) {
		return withGoReleases(func(r []string) (string, error) { return goNewestIn(r, line) })
	},
	Support: func(line string) error {
		_, err := withGoReleases(func(r []string) (string, error) { return "", goSupport(r, line) })
		return err
	},
}

const goPolicy = "https://go.dev/doc/devel/release#policy"

var goLine = regexp.MustCompile(`^1\.([0-9]+)$`)

// withGoReleases reads the stable Go releases from go.dev (for example
// "1.27.3") and passes them to f.
func withGoReleases(f func([]string) (string, error)) (string, error) {
	data, err := sys.Get("https://go.dev/dl/?mode=json&include=all")
	if err != nil {
		return "", err
	}
	releases, err := parseGoReleases(data)
	if err != nil {
		return "", err
	}
	return f(releases)
}

// parseGoReleases returns the stable releases of the go.dev release list,
// without the prefix "go".
func parseGoReleases(data []byte) ([]string, error) {
	var list []struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}
	var releases []string
	for _, r := range list {
		if r.Stable {
			releases = append(releases, strings.TrimPrefix(r.Version, "go"))
		}
	}
	if len(releases) == 0 {
		return nil, fmt.Errorf("no stable Go release on go.dev")
	}
	return releases, nil
}

// goMinor returns N of a Go line "1.N" (-1 when the line is not valid).
func goMinor(line string) int {
	m := goLine.FindStringSubmatch(line)
	if m == nil {
		return -1
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

// goNewestIn returns the newest release of the line, for example "1.27.3"
// for "1.27".
func goNewestIn(releases []string, line string) (string, error) {
	if goMinor(line) < 0 {
		return "", fmt.Errorf("GO_PIN=%s: a Go line has the form 1.N, for example 1.27", line)
	}
	best := ""
	for _, r := range releases {
		if layer.InLine(r, line) && (best == "" || compareVersions(r, best) > 0) {
			best = r
		}
	}
	if best == "" {
		return "", fmt.Errorf("Go %s has no stable release", line)
	}
	return best, nil
}

// goSupport returns an *layer.EndOfLifeError when Go 1.(N+2) is released.
func goSupport(releases []string, line string) error {
	n := goMinor(line)
	if n < 0 {
		return fmt.Errorf("GO_PIN=%s: a Go line has the form 1.N, for example 1.27", line)
	}
	newest := -1
	for _, r := range releases {
		parts := strings.SplitN(r, ".", 3)
		if len(parts) >= 2 {
			if m := goMinor(parts[0] + "." + parts[1]); m > newest {
				newest = m
			}
		}
	}
	switch {
	case n > newest:
		return fmt.Errorf("Go %s is not released yet (newest: 1.%d)", line, newest)
	case n >= newest-1:
		return nil
	}
	return &layer.EndOfLifeError{
		Since:     fmt.Sprintf("Go 1.%d.0 was released; Go supports the two newest major releases", n+2),
		Source:    goPolicy,
		Supported: []string{fmt.Sprintf("1.%d", newest-1), fmt.Sprintf("1.%d", newest)},
	}
}

// goModuleFor returns the newest release of a Go module whose go.mod needs at
// most the Go version goVersion, from the Go module proxy. The Go tools are
// built with the installed Go (GOTOOLCHAIN=local), so they must accept it.
func goModuleFor(module, goVersion string) (string, error) {
	base := "https://proxy.golang.org/" + strings.ToLower(module) + "/@v/"
	data, err := sys.Get(base + "list")
	if err != nil {
		return "", err
	}
	versions := releaseVersions(string(data))
	const maxTries = 30
	for i, v := range versions {
		if i == maxTries {
			break
		}
		mod, err := sys.Get(base + v + ".mod")
		if err != nil {
			return "", err
		}
		if compareVersions(goDirective(string(mod)), goVersion) <= 0 {
			return v, nil
		}
	}
	return "", fmt.Errorf("no release of %s in the newest %d works with Go %s", module, maxTries, goVersion)
}

// releaseVersions returns the versions vX.Y.Z of a module proxy list without
// pre-releases, newest first.
func releaseVersions(list string) []string {
	release := regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)
	var versions []string
	for _, v := range strings.Fields(list) {
		if release.MatchString(v) {
			versions = append(versions, v)
		}
	}
	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) > 0 })
	return versions
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

// compareVersions compares two versions by their numbers (the prefixes "v"
// and "go" do not count; missing numbers count as 0): -1, 0 or 1.
func compareVersions(a, b string) int {
	numbers := func(v string) []int {
		v = strings.TrimPrefix(strings.TrimPrefix(v, "go"), "v")
		var result []int
		for _, part := range strings.Split(v, ".") {
			digits := regexp.MustCompile(`^[0-9]+`).FindString(part)
			n, _ := strconv.Atoi(digits)
			result = append(result, n)
		}
		return result
	}
	x, y := numbers(a), numbers(b)
	for i := 0; i < len(x) || i < len(y); i++ {
		var p, q int
		if i < len(x) {
			p = x[i]
		}
		if i < len(y) {
			q = y[i]
		}
		if p != q {
			if p < q {
				return -1
			}
			return 1
		}
	}
	return 0
}
