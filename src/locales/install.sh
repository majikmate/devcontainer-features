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
update-locale LANG="$LANG" LC_TIME="$TIME" LC_NUMERIC="$NUMERIC" LC_MONETARY="$MONETARY" LC_MEASUREMENT="$MEASUREMENT"

ln -fs "/usr/share/zoneinfo/${TIMEZONE}" /etc/localtime
dpkg-reconfigure -f noninteractive tzdata

# The 'install.sh' entrypoint script is always executed as the root user.
#
# These following environment variables are passed in by the dev container CLI.
# These may be useful in instances where the context of the final 
# remoteUser or containerUser is useful.
# For more details, see https://containers.dev/implementors/features#user-env-var

# echo "The effective dev container remoteUser is '$_REMOTE_USER'"
# echo "The effective dev container remoteUser's home directory is '$_REMOTE_USER_HOME'"

# echo "The effective dev container containerUser is '$_CONTAINER_USER'"
# echo "The effective dev container containerUser's home directory is '$_CONTAINER_USER_HOME'"

# cat > /usr/local/bin/color \
# << EOF
# #!/bin/sh
# echo "my favorite color is ${FAVORITE}"
# EOF

# chmod +x /usr/local/bin/color

