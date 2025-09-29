#!/usr/bin/env bash
set -euo pipefail

# --- config (can be overridden via flags or env) ---
NAMESPACE="${NAMESPACE:-majikmate/devcontainer-features}"
SOURCE="${SOURCE:-ghcr.io}"
PROJECT_ROOT="${PROJECT_ROOT:-.}"
LOG_LEVEL="${LOG_LEVEL:-info}"
GITHUB_OWNER="${GITHUB_OWNER:-}"
GITHUB_REPO="${GITHUB_REPO:-}"

# --- helpers ---
die() { echo "❌ $*" >&2; exit 1; }
log() { echo "▶ $*"; }

need() { command -v "$1" >/dev/null 2>&1 || die "Missing required tool: $1"; }

usage() {
  cat >&2 <<EOF
Usage: $0 [options]

Options:
  --namespace <owner/repo>     OCI namespace for features (default: $NAMESPACE)
  --source <registry>          OCI registry (default: $SOURCE)
  --project-root <path>        Repo root (default: $PROJECT_ROOT)
  --github-owner <owner>       GitHub owner (optional; inferred from 'origin' if unset)
  --github-repo  <repo>        GitHub repo  (optional; inferred from 'origin' if unset)
  --log-level <level>          devcontainer log level: info|debug|trace (default: $LOG_LEVEL)
  -h, --help                   Show help

Environment overrides are supported for all options (NAMESPACE, SOURCE, PROJECT_ROOT, GITHUB_OWNER, GITHUB_REPO, LOG_LEVEL).
EOF
}

# --- arg parse ---
while [[ $# -gt 0 ]]; do
  case "$1" in
    --namespace) NAMESPACE="$2"; shift 2;;
    --source) SOURCE="$2"; shift 2;;
    --project-root) PROJECT_ROOT="$2"; shift 2;;
    --github-owner) GITHUB_OWNER="$2"; shift 2;;
    --github-repo) GITHUB_REPO="$2"; shift 2;;
    --log-level) LOG_LEVEL="$2"; shift 2;;
    -h|--help) usage; exit 0;;
    *) die "Unknown option: $1 (use --help)";;
  esac
done

# --- checks ---
need devcontainer
need jq
[ -d "$PROJECT_ROOT/src" ] || die "No 'src/' folder under $PROJECT_ROOT"

# --- infer GitHub owner/repo if missing ---
if [[ -z "$GITHUB_OWNER" || -z "$GITHUB_REPO" ]]; then
  if git -C "$PROJECT_ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    origin_url="$(git -C "$PROJECT_ROOT" remote get-url origin 2>/dev/null || true)"
    if [[ -n "$origin_url" ]]; then
      # supports git@github.com:owner/repo.git or https://github.com/owner/repo(.git)
      if [[ "$origin_url" =~ github.com[:/]+([^/]+)/([^/.]+) ]]; then
        GITHUB_OWNER="${GITHUB_OWNER:-${BASH_REMATCH[1]}}"
        GITHUB_REPO="${GITHUB_REPO:-${BASH_REMATCH[2]}}"
      fi
    fi
  fi
fi

log "Namespace:        $NAMESPACE"
log "Registry source:  $SOURCE"
log "Project root:     $PROJECT_ROOT"
log "GitHub owner:     ${GITHUB_OWNER:-<unset>}"
log "GitHub repo:      ${GITHUB_REPO:-<unset>}"
log "Log level:        $LOG_LEVEL"

# --- discover features ---
FEATURES=()
for dir in "$PROJECT_ROOT/src"/*; do
    if [ -d "$dir" ] && [ -f "$dir/devcontainer-feature.json" ]; then
        FEATURES+=("$(basename "$dir")")
    fi
done

# Sort the features array
if [ ${#FEATURES[@]} -gt 0 ]; then
    IFS=$'\n' FEATURES=($(sort <<<"${FEATURES[*]}"))
    unset IFS
fi

[ ${#FEATURES[@]} -gt 0 ] || die "No features found (expected folders with devcontainer-feature.json under $PROJECT_ROOT/src)"

log "Found ${#FEATURES[@]} feature(s): ${FEATURES[*]}"

# --- generate per-feature READMEs ---
log "Generating feature READMEs under src/…"
devcontainer features generate-docs \
  --project-folder "$PROJECT_ROOT/src" \
  --namespace "$NAMESPACE" \
  ${GITHUB_OWNER:+--github-owner "$GITHUB_OWNER"} \
  ${GITHUB_REPO:+--github-repo "$GITHUB_REPO"} \
  --log-level "$LOG_LEVEL"

# --- generate root README (comprehensive feature catalog) ---
log "Generating root README with all features…"

# Build features table from devcontainer-feature.json files
FEATURES_TABLE=""
for feature in "${FEATURES[@]}"; do
    feature_json="$PROJECT_ROOT/src/$feature/devcontainer-feature.json"
    if [ -f "$feature_json" ]; then
        name=$(jq -r '.name // "N/A"' "$feature_json")
        description=$(jq -r '.description // "N/A"' "$feature_json")
        FEATURES_TABLE+="| [$feature](./src/$feature) | $name | $description |\n"
    fi
done

# Generate the root README
cat > "$PROJECT_ROOT/README.md" <<EOF
# DevContainer Features

This repository contains a collection of
[Dev Container features](https://containers.dev/implementors/features/) for enhanced
development environment setup and configuration.

## Usage

To use a feature, add it to your \`devcontainer.json\` file:

\`\`\`json
{
    "features": {
        "ghcr.io/$NAMESPACE/feature-name:1": {}
    }
}
\`\`\`

## Available Features

| Feature | Name | Description |
|---------|------|-------------|
$(printf "%b" "$FEATURES_TABLE")

## Examples

### Basic Usage

\`\`\`json
{
    "features": {
$(for feature in "${FEATURES[@]}"; do
    printf '        "ghcr.io/%s/%s:1": {},\n' "$NAMESPACE" "$feature"
done | sed '$ s/,$//')
    }
}
\`\`\`

### Advanced Configuration

\`\`\`json
{
    "features": {
$(for feature in "${FEATURES[@]}"; do
    feature_json="$PROJECT_ROOT/src/$feature/devcontainer-feature.json"
    if [ -f "$feature_json" ]; then
        printf '        "ghcr.io/%s/%s:1": {\n' "$NAMESPACE" "$feature"
        # Add sample options if they exist
        options=$(jq -r '.options // {} | keys[]' "$feature_json" 2>/dev/null | head -2)
        if [ -n "$options" ]; then
            while IFS= read -r option; do
                default_value=$(jq -r ".options[\"$option\"].default // \"\"" "$feature_json")
                printf '            "%s": "%s",\n' "$option" "$default_value"
            done <<< "$options" | sed '$ s/,$//'
        fi
        printf '        },\n'
    fi
done | sed '$ s/,$//')
    }
}
\`\`\`

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](./LICENSE).
EOF

log "Wrote $PROJECT_ROOT/README.md"

log "✅ Done."