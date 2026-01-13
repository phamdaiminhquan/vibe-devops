#!/bin/sh

# install.sh: A script to install the 'vibe' CLI from GitHub Releases.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/phamdaiminhquan/vibe-devops/main/install.sh | sh
#
# This script will automatically:
# 1. Detect the user's OS and architecture.
# 2. Fetch the latest release from the vibe-devops GitHub repository.
# 3. Download the correct binary for the detected platform.
# 4. Unpack it and move it to /usr/local/bin.

set -e

# Define the GitHub repository
REPO="phamdaiminhquan/vibe-devops"
BINARY_NAME="vibe"
INSTALL_DIR="/usr/local/bin"

# --- Helper Functions ---
print_info() {
    printf "\033[1;36m%s\033[0m\n" "$1"
}

print_error() {
    printf "\033[1;31mError: %s\033[0m\n" "$1" >&2
    exit 1
}

# --- Dependency Check ---
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

if ! command_exists curl; then
    print_error "curl is not installed. Please install it first."
fi


# --- Platform Detection ---
get_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux";;
        Darwin*) echo "darwin";;
        *)       print_error "Unsupported operating system: $(uname -s)";;
    esac
}

get_arch() {
    case "$(uname -m)" in
        x86_64)  echo "amd64";;
        arm64)   echo "arm64";;
        aarch64) echo "arm64";; # aarch64 is common on Linux arm64 systems
        *)       print_error "Unsupported architecture: $(uname -m)";;
    esac
}

# --- Main Logic ---
main() {
    OS=$(get_os)
    ARCH=$(get_arch)
    print_info "Detected Platform: ${OS}/${ARCH}"

    print_info "Fetching latest release from GitHub..."
    LATEST_RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
    
    if command_exists jq; then
        DOWNLOAD_URL=$(curl -sSL "$LATEST_RELEASE_URL" | jq -r \
            ".assets[] | select(.name | test(\"${OS}_${ARCH}\")) | .browser_download_url")
    else
        # Fallback to grep/cut if jq is not available
        DOWNLOAD_URL=$(curl -sSL "$LATEST_RELEASE_URL" | grep "browser_download_url" | grep "${OS}_${ARCH}" | cut -d '"' -f 4 | head -n 1)
    fi

    if [ -z "$DOWNLOAD_URL" ]; then
        print_error "Could not find a download for your platform (${OS}/${ARCH})."
    fi

    print_info "Downloading from: ${DOWNLOAD_URL}" 
    
    # Create a temporary directory for the download
    TMP_DIR=$(mktemp -d)
    trap 'rm -rf -- "$TMP_DIR"' EXIT # Clean up on exit

    ARCHIVE_NAME=$(basename "$DOWNLOAD_URL")
    curl -L --progress-bar "$DOWNLOAD_URL" -o "${TMP_DIR}/${ARCHIVE_NAME}"

    print_info "Unpacking archive..."
    if [ "${OS}" = "linux" ] || [ "${OS}" = "darwin" ]; then
        tar -xzf "${TMP_DIR}/${ARCHIVE_NAME}" -C "$TMP_DIR"
    else
        print_error "Unsupported OS for unpacking."
    fi

    print_info "Installing '${BINARY_NAME}' to ${INSTALL_DIR}..."
    
    if [ -w "$INSTALL_DIR" ]; then
        # User has write permissions
        mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        # User does not have write permissions, try with sudo
        print_info "Write permission to ${INSTALL_DIR} is required. Trying with sudo..."
        sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    print_info "✅ '${BINARY_NAME}' has been installed successfully!"
    print_info "You can now run 'vibe --help' to get started."
}

# --- Run ---
main
