#!/usr/bin/env sh

set -eu

REPO_OWNER="sandevgo"
REPO_NAME="tuskbot"
BINARY_NAME="tusk"

say() {
  printf '%s\n' "$*"
}

fail() {
  printf 'Error: %s\n' "$*" >&2
  exit 1
}

download() {
  url="$1"
  output="$2"

  if command -v curl >/dev/null 2>&1; then
    curl -fL "$url" -o "$output"
    return 0
  fi

  if command -v wget >/dev/null 2>&1; then
    wget -qO "$output" "$url"
    return 0
  fi

  fail "Neither curl nor wget is installed. Install one of them and retry."
}

if ! command -v tar >/dev/null 2>&1; then
  fail "tar is required but not installed."
fi

case "$(uname -s)" in
  Linux)
    os="linux"
    ;;
  Darwin)
    os="darwin"
    ;;
  *)
    fail "Unsupported OS: $(uname -s). Supported: Linux and macOS."
    ;;
esac

case "$(uname -m)" in
  x86_64 | amd64)
    arch="amd64"
    ;;
  arm64 | aarch64)
    arch="arm64"
    ;;
  *)
    fail "Unsupported architecture: $(uname -m). Supported: amd64 and arm64."
    ;;
esac

if [ "$os" = "darwin" ] && [ "$arch" != "arm64" ]; then
  fail "No macOS $arch release artifact is available. Supported macOS architecture: arm64."
fi

asset="tusk-${os}-${arch}.tar.gz"
release_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/${asset}"

tmp_dir="$(mktemp -d 2>/dev/null || mktemp -d -t tusk-install)"
cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

archive_path="$tmp_dir/$asset"
extract_dir="$tmp_dir/extracted"

say "Downloading ${asset} from latest stable release..."
download "$release_url" "$archive_path"

say "Extracting archive..."
mkdir -p "$extract_dir"
tar -xzf "$archive_path" -C "$extract_dir"

candidate_1="$extract_dir/bin/tusk-${os}-${arch}"
candidate_2="$extract_dir/tusk-${os}-${arch}"
candidate_3="$extract_dir/tusk"

if [ -f "$candidate_1" ]; then
  binary_source="$candidate_1"
elif [ -f "$candidate_2" ]; then
  binary_source="$candidate_2"
elif [ -f "$candidate_3" ]; then
  binary_source="$candidate_3"
else
  fail "Could not locate the tusk binary in ${asset}."
fi

if [ -n "${TUSK_INSTALL_DIR:-}" ]; then
  install_dir="$TUSK_INSTALL_DIR"
elif [ -n "${XDG_BIN_HOME:-}" ]; then
  install_dir="$XDG_BIN_HOME"
else
  if [ -z "${HOME:-}" ]; then
    fail "HOME is not set. Set TUSK_INSTALL_DIR to continue."
  fi

  case ":${PATH:-}:" in
    *":$HOME/.local/bin:"*)
      install_dir="$HOME/.local/bin"
      ;;
    *":$HOME/bin:"*)
      install_dir="$HOME/bin"
      ;;
    *)
      install_dir="$HOME/.local/bin"
      ;;
  esac
fi

mkdir -p "$install_dir"
install_path="$install_dir/$BINARY_NAME"

say "Installing ${BINARY_NAME} to ${install_path}..."
cp "$binary_source" "$install_path"
chmod +x "$install_path"

say "Verifying binary..."
"$install_path" version >/dev/null

case ":$PATH:" in
  *":$install_dir:"*)
    ;;
  *)
    say ""
    say "Warning: ${install_dir} is not in your PATH."
    say "Add this to ~/.zshrc or ~/.bashrc and restart your shell:"
    say "  export PATH=\"${install_dir}:\$PATH\""
    say ""
    ;;
esac

say "Running interactive setup: tusk install"
if [ -r /dev/tty ]; then
  "$install_path" install </dev/tty
else
  fail "Interactive setup requires a TTY. Run '${install_path} install' manually."
fi

say "Installing service: tusk service install"
"$install_path" service install

say "Starting service: tusk service start"
"$install_path" service start

say ""
say "Done. Check status with: ${install_path} service status"
