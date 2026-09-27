// Package features is the collection of distribution-independent layers of
// the Dev Container images: git, aliases, pure-prompt, go, node, deno,
// prettier and github-cli.
//
// A feature downloads, checks and configures; it does not use the package
// manager of the distribution and no distribution-specific paths. What it
// needs from the system, it declares as a needed layer of devcontainer-core
// (for example node needs build-tools). The framework (layer definition,
// helpers, VS Code metadata) is in devcontainer-core; devcontainer-core also
// builds the program devenv with these features.
//
// Each file registers one layer in its init function; importing the package
// registers all of them:
//
//	import _ "github.com/majikmate/devcontainer-features"
package features
