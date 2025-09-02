#!/bin/sh
set -e

echo "Activating feature 'updateos'"

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


cat > /usr/local/share/majikmate-update-os << 'EOF'
#!/bin/sh
echo "starting OS update."
apt-get update
apt-get upgrade -y
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/*
EOF

chmod +x /usr/local/share/majikmate-update-os


# create the atcreate script
if [ "${ATCREATE}" = "true" ]; then
	cat > /usr/local/share/majikmate-update-os-at-create <<EOF
#!/bin/sh
sudo /usr/local/share/majikmate-update-os
EOF
else
	cat > /usr/local/share/majikmate-update-os-at-create <<EOF
#!/bin/sh
echo "OS update at create skipped."
EOF
fi

chmod +x /usr/local/share/majikmate-update-os-at-create


# create the atstart script
if [ "${ATSTART}" = "true" ]; then
	cat > /usr/local/share/majikmate-update-os-at-start <<EOF
#!/bin/sh
sudo /usr/local/share/majikmate-update-os
EOF
else
	cat > /usr/local/share/majikmate-update-os-at-start <<EOF
#!/bin/sh
echo "OS update at start skipped."
EOF
fi

chmod +x /usr/local/share/majikmate-update-os-at-start

