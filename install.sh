#!/bin/bash

set -e

REPO="infraspecdev/mini-heroku"
BINARY="mini"

OS=$(uname -s)

if [ "$OS" = "Linux"]; then
    FILE="mini-linux-amd64"
elif [ "$OS" = "Darwin"]; then
    FILE="mini-darwin-amd64"
else
    echo "Unsupported OS"
    exit 1
fi

URL="https://github.com/$REPO/releases/latest/download/$FILE"

echo "Downloading mini CLI..."

curl -L $URL -o $BINARY

chmod +x $BINARY

sudo mv $BINARY /usr/local/bin

echo "mini installed successfully!"
