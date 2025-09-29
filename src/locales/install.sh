#!/bin/sh
set -e

echo "Activating feature 'locales'"
echo "============================"
echo "timezone: ${TIMEZONE}"
echo "language: ${LANG}"
echo "time format: ${TIME}"
echo "numeric format: ${NUMERIC}"
echo "monetary format: ${MONETARY}"
echo "measurement format: ${MEASUREMENT}"

apt-get update
apt-get install -y tzdata
apt-get install -y locales

ensure_locale() {
    _locale=$1
    if [ -z "${_locale}" ]; then
        echo "ensure_locale: No locale provided."
        return 1
    fi

    _locale_underscore=$(echo "${_locale}" | sed 's/-/_/')
    _locale_entry="${_locale_underscore} UTF-8"

    if grep -q -E "^\s*${_locale_entry}\s*$" /etc/locale.gen; then
        echo "Locale ${_locale_underscore} already enabled."
    elif grep -q -E "^\s*#\s*${_locale_entry}\s*$" /etc/locale.gen; then
        echo "Enabling locale ${_locale_underscore}."
        sed -i -E "s/^\s*#\s*${_locale_entry}\s*$/${_locale_entry}/" /etc/locale.gen
    else
        echo "Adding locale ${_locale_underscore} to /etc/locale.gen."
        echo "${_locale_entry}" >> /etc/locale.gen
    fi
}

ensure_locale "$LANG"
ensure_locale "$TIME"
ensure_locale "$NUMERIC"
ensure_locale "$MONETARY"
ensure_locale "$MEASUREMENT"

dpkg-reconfigure -f noninteractive locales

# Construct the LANGUAGE variable from LANG, e.g., en_GB.UTF-8 -> en_GB:en
LANG_UNDERSCORE=$(echo "$LANG" | sed 's/\.UTF-8//' | sed 's/-/_/')
LANG_SHORT=$(echo "$LANG_UNDERSCORE" | cut -d'_' -f1)
LANGUAGE_VALUE="${LANG_UNDERSCORE}:${LANG_SHORT}"

update-locale LANG="${LANG}" LANGUAGE="${LANGUAGE_VALUE}" LC_TIME="${TIME}" LC_NUMERIC="${NUMERIC}" LC_MONETARY="${MONETARY}" LC_MEASUREMENT="${MEASUREMENT}"

# Resolve users
REMOTE_USER="${_REMOTE_USER:-$(id -un 2>/dev/null || true)}"
CONTAINER_USER="${_CONTAINER_USER:-root}"

# Try to resolve remote user home from the env var, passwd, or a sane default
resolve_home() {
    u="$1"
    h="${2:-}"
    if [ -n "$h" ]; then
        printf '%s' "$h"; return 0
    fi
    # If user exists in /etc/passwd, use its home
    if getent passwd "$u" >/dev/null 2>&1; then
        getent passwd "$u" | cut -d: -f6
        return 0
    fi
    # Common default
    if [ -d "/home/$u" ]; then
        printf '/home/%s' "$u"; return 0
    fi
    # Last resort (build-time before user exists)
    printf '%s' "${_CONTAINER_USER_HOME:-/root}"
}

REMOTE_USER_HOME="$(resolve_home "$REMOTE_USER" "${_REMOTE_USER_HOME:-}")"
CONTAINER_USER_HOME="$(resolve_home "$CONTAINER_USER" "${_CONTAINER_USER_HOME:-}")"

echo "The effective dev container remoteUser is '${REMOTE_USER}'"
echo "The effective dev container remoteUser's home directory is '${REMOTE_USER_HOME}'"
echo "The effective dev container containerUser is '${CONTAINER_USER}'"
echo "The effective dev container containerUser's home directory is '${CONTAINER_USER_HOME}'"

# Force sourcing of /etc/default/locale for the remoteUser's shell

# Add locale export to shell configuration files
echo "Adding locale configuration to shell configuration files"

# Prepare locale content to add
LOCALE_CONFIG="# Load system-wide locale settings\nexport \$(grep -v '^#' /etc/default/locale | xargs)"

append_if_exists() {
    target="$1"
    if [ -f "$target" ]; then
        echo "Adding locale configuration to $target"
        [ -n "$LOCALE_CONFIG" ] && printf "%b\n" "$LOCALE_CONFIG" >> "$target"
    else
        echo "$target not found, skipping"
    fi
}

# Prefer drop-in files that work even before the user exists
# Bash (Debian/Ubuntu): /etc/profile.d
if [ -n "$LOCALE_CONFIG" ]; then
    echo "Installing shell-agnostic locale snippets"
    printf "%b\n" "$LOCALE_CONFIG" > /etc/profile.d/99-locales.sh
    chmod 0644 /etc/profile.d/99-locales.sh

    # Zsh global config locations (use ones that exist)
    if [ -d /etc/zsh/zshrc.d ]; then
        printf "%b\n" "$LOCALE_CONFIG" > /etc/zsh/zshrc.d/99-locales.zsh
        chmod 0644 /etc/zsh/zshrc.d/99-locales.zsh
    elif [ -f /etc/zsh/zshrc ]; then
        printf "\n%b\n" "$LOCALE_CONFIG" >> /etc/zsh/zshrc
    fi
fi

# Also try per-user shells if the home exists at this point
append_if_exists "${REMOTE_USER_HOME}/.bashrc"
append_if_exists "${REMOTE_USER_HOME}/.zshrc"

ln -fs "/usr/share/zoneinfo/${TIMEZONE}" /etc/localtime
dpkg-reconfigure -f noninteractive tzdata