#!/bin/bash

set -e

# Detect operating system
OS_NAME="$(uname -s)"
case "${OS_NAME}" in
    Linux*)     OS=linux;;
    Darwin*)    OS=darwin;;
    *)
        echo "Unsupported operating system: ${OS_NAME}"
        exit 1
        ;;
esac

ARCH_NAME="$(uname -m)"
case "${ARCH_NAME}" in
    x86_64*)    ARCH_NAME=amd64;;
    arm64*)     ARCH_NAME=arm64;;
    armv8*)     ARCH_NAME=arm64;;
    armv7*)     ARCH_NAME=arm;;
    *)
        echo "Unsupported architecture: ${ARCH_NAME}"
        exit 1
        ;;
esac

# Creating .local/bin if it does not exist
LOCAL_BIN="$HOME/.local/bin"
if [ ! -d "${LOCAL_BIN}" ]; then
    mkdir -p "${LOCAL_BIN}"
fi

# If .local/bin is not in PATH, add it
if [[ ":$PATH:" != *":$LOCAL_BIN:"* ]]; then
  if [[ $SHELL == *"bash"* ]]; then
    echo 'export PATH=$PATH:$HOME/.local/bin' >> ~/.bashrc
    source ~/.bashrc
  elif [[ $SHELL == *"zsh"* ]]; then
    echo 'export PATH=$PATH:$HOME/.local/bin' >> ~/.zshrc
    source ~/.zshrc
  elif [[ $SHELL == *"fish"* ]]; then
    echo 'set PATH $PATH $HOME/.local/bin' > ~/.config/fish/conf.d/bulut.fish
    source ~/.config/fish/conf.d/bulut.fish
  else
    echo "Unsupported shell: ${SHELL}"
    echo "Please add ${LOCAL_BIN} to your PATH manually."
  fi
fi

# Downloading the binary from the GitHub release page
TEMP_DIR="/tmp/bulut"
mkdir -p "${TEMP_DIR}"
BULUT_ARCHIVE="${TEMP_DIR}/bulut.tar.gz"
BULUT_BIN="${TEMP_DIR}/bulut"
REPO="unitythemaker/bulut"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"
# Fetch the release information from GitHub API
RELEASE_INFO=$(curl -s "${API_URL}")
# Get the tag name from the latest release
TAG_NAME=$(echo "${RELEASE_INFO}" | grep "tag_name" | cut -d "\"" -f 4)
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG_NAME}/bulut-${TAG_NAME}-${OS}-${ARCH_NAME}.tar.gz"

echo $DOWNLOAD_URL

# if there is no download url for the current OS
if [ "${DOWNLOAD_URL}" == "" ]; then
    echo "No bulut binary available for ${OS}-${ARCH_NAME} version ${TAG_NAME}"
    exit 1
fi

echo "Downloading the latest version of bulut (${TAG_NAME})..."
curl -L -o "${BULUT_ARCHIVE}" "${DOWNLOAD_URL}"

# Extracting the binary
tar -xzf "${BULUT_ARCHIVE}" -C "${TEMP_DIR}"

# Moving binary to .local/bin
chmod a+x "${BULUT_BIN}"
mv "${BULUT_BIN}" "${LOCAL_BIN}/bulut"

# Cleaning up
rm "${BULUT_ARCHIVE}"
rm -rf "/tmp/bulut"

echo "Bulut has been installed in ${LOCAL_BIN}"
echo "If you have not added ${LOCAL_BIN} to your PATH, please do so."
echo "We already added it to your .bashrc file."
