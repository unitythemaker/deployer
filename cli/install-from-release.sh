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

# Creating .local/bin if it does not exist
LOCAL_BIN="$HOME/.local/bin"
if [ ! -d "${LOCAL_BIN}" ]; then
    mkdir -p "${LOCAL_BIN}"
fi

# If .local/bin is not in PATH, add it
if [[ $SHELL == *"bash"* ]] && [[ ":$PATH:" != *":$LOCAL_BIN:"* ]]; then
    echo 'export PATH=$PATH:$HOME/.local/bin' >> ~/.bashrc
    source ~/.bashrc
fi

# Downloading the binary from the GitHub release page
TEMP_DIR="/tmp/deployer"
mkdir -p "${TEMP_DIR}"
DEPLOYER_ZIP="${TEMP_DIR}/deployer.zip"
DEPLOYER_BIN="${TEMP_DIR}/deployer"
REPO="unitythemaker/deployer"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"
# Fetch the release information from GitHub API
RELEASE_INFO=$(curl -s "${API_URL}")
# Get the tag name from the latest release
TAG_NAME=$(echo "${RELEASE_INFO}" | grep "tag_name" | cut -d "\"" -f 4)
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG_NAME}/deployer-${OS}-${ARCH_NAME}.zip"

# if there is no download url for the current OS
if [ "${DOWNLOAD_URL}" == "" ]; then
    echo "No deployer binary available for ${OS}-${ARCH_NAME} version ${TAG_NAME}"
    exit 1
fi

echo "Downloading the latest version of deployer (${TAG_NAME})..."
curl -L -o "${DEPLOYER_ZIP}" "${DOWNLOAD_URL}"

# Extracting the binary
unzip -q "${DEPLOYER_ZIP}" -d "${TEMP_DIR}"

# Moving binary to .local/bin
chmod a+x "${DEPLOYER_BIN}"
mv "${DEPLOYER_BIN}" "${LOCAL_BIN}/deployer"

# Cleaning up
rm "${DEPLOYER_ZIP}"
rm -rf "/tmp/deployer"

echo "Deployer has been installed in ${LOCAL_BIN}"
echo "If you have not added ${LOCAL_BIN} to your PATH, please do so."
echo "We already added it to your .bashrc file."
