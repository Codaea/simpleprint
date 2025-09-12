#!/bin/bash
set -e

# Configuration
REPO="codaea/simpleprint"
SERVICE_NAME="simpleprint"
INSTALL_DIR="/usr/bin"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
WORK_DIR="/tmp/simpleprint"

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
	x86_64)
		ARCH=amd64
		;;
	aarch64)
		ARCH=arm64
		;;
	armv7l)
		ARCH=armv7
		;;
    armv6l)
        ARCH=armv6
        ;;
	*)
		echo "Unsupported architecture: $ARCH"
		exit 1
		;;
esac

# Get latest release tag
TAG=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep 'tag_name' | cut -d '"' -f4)
if [ -z "$TAG" ]; then
	echo "Failed to fetch latest release tag."
	exit 1
fi


# Download binary
BINARY_URL="https://github.com/${REPO}/releases/download/${TAG}/${SERVICE_NAME}_${TAG#v}_linux_${ARCH}"
TMP_BIN="/tmp/${SERVICE_NAME}"
echo "Downloading $BINARY_URL ..."
curl -L "$BINARY_URL" -o "$TMP_BIN"
chmod +x "$TMP_BIN"

# If binary exists, backup and replace
if [ -f "$INSTALL_DIR/$SERVICE_NAME" ]; then
	echo "Existing binary detected at $INSTALL_DIR/$SERVICE_NAME. Updating..."
	sudo mv "$INSTALL_DIR/$SERVICE_NAME" "$INSTALL_DIR/${SERVICE_NAME}.bak.$(date +%s)"
fi
sudo mv "$TMP_BIN" "$INSTALL_DIR/$SERVICE_NAME"
sudo chmod +x "$INSTALL_DIR/$SERVICE_NAME"

# Ensure service file exists, download if missing
SERVICE_SRC="$WORK_DIR/sample.service"
if [ ! -f "$SERVICE_SRC" ]; then
	echo "sample.service not found locally. Downloading from GitHub..."
	RAW_URL="https://raw.githubusercontent.com/${REPO}/${TAG}/sample.service"
	curl -L "$RAW_URL" -o "$SERVICE_SRC"
fi

sudo cp "$SERVICE_SRC" "/etc/systemd/system/simpleprint.service"
sudo systemctl daemon-reload
sudo systemctl enable "$SERVICE_NAME"
sudo systemctl restart "$SERVICE_NAME"

# Ensure .env file exists in /home/coda/simpleprint

# Use ~/.simpleprint/.env for environment
ENV_DIR="$HOME/.simpleprint"
ENV_FILE="$ENV_DIR/.env"
if [ ! -d "$ENV_DIR" ]; then
	mkdir -p "$ENV_DIR"
fi
if [ ! -f "$ENV_FILE" ]; then
	echo ".env not found in $ENV_DIR. Creating a default .env file..."
	RAW_ENV_URL="https://raw.githubusercontent.com/${REPO}/${TAG}/lib/example.env"
	if curl --output /dev/null --silent --head --fail "$RAW_ENV_URL"; then
		curl -L "$RAW_ENV_URL" -o "$ENV_FILE"
	else
		cat <<EOF > "$ENV_FILE"
# Default environment for simpleprint
PRINTER_DEVICE=/dev/usb/lp0
LOG_LEVEL=info
PORT=8080
EOF
	fi
fi

# Update systemd service to use new env file location
sudo sed -i "s|^EnvironmentFile=.*|EnvironmentFile=$ENV_FILE|" /etc/systemd/system/simpleprint.service

echo "Installation complete. Service status:"
systemctl status "$SERVICE_NAME" --no-pager
