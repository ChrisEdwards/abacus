# Installation Guide

This guide covers various methods for installing Abacus on different platforms.

## Prerequisites

### Required

- **Beads CLI v0.24.0 or later** - Install from [github.com/steveyegge/beads](https://github.com/steveyegge/beads) and verify with `bd --version`

### Recommended

- Terminal with 256-color support
- Go 1.25.3 or later (only required for `go install` or building from source)

### Verify Prerequisites

```bash
# Check Beads installation
bd --version

# Check terminal color support
echo $TERM

# Check Go version (if using go install or building from source)
go version
```

## Installation Methods

### Method 1: Homebrew (Recommended)

The easiest way to install Abacus on macOS or Linux is using Homebrew:

```bash
brew tap ChrisEdwards/tap
brew install abacus
```

This will:
- Add the ChrisEdwards/tap to your Homebrew taps
- Download the pre-built binary for your platform
- Install it to Homebrew's bin directory
- Automatically add it to your PATH

**Verify installation:**
```bash
abacus --version
```

**Update Abacus:**
```bash
brew upgrade abacus
```

**Uninstall:**
```bash
brew uninstall abacus
```

### Method 2: Install Script

Use the automated install script for Unix/macOS/Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/ChrisEdwards/abacus/main/scripts/install.sh | bash
```

This script will:
- Detect your platform and architecture automatically
- Download the latest release binary
- Install to `~/.local/bin` (configurable via `INSTALL_DIR`)
- Provide PATH setup instructions if needed

**Alternative installation directory:**
```bash
INSTALL_DIR=/usr/local/bin curl -fsSL https://raw.githubusercontent.com/ChrisEdwards/abacus/main/scripts/install.sh | bash
```

### Method 3: Direct Download

Download pre-built binaries from [GitHub Releases](https://github.com/ChrisEdwards/abacus/releases):

1. Visit the [latest release page](https://github.com/ChrisEdwards/abacus/releases/latest)
2. Download the appropriate archive for your platform:
   - `abacus_X.Y.Z_darwin_amd64.tar.gz` - macOS Intel
   - `abacus_X.Y.Z_darwin_arm64.tar.gz` - macOS Apple Silicon
   - `abacus_X.Y.Z_linux_amd64.tar.gz` - Linux x64
   - `abacus_X.Y.Z_linux_arm64.tar.gz` - Linux ARM64
   - `abacus_X.Y.Z_windows_amd64.tar.gz` - Windows

3. Extract and install:

**macOS/Linux:**
```bash
tar -xzf abacus_*.tar.gz
sudo mv abacus /usr/local/bin/
chmod +x /usr/local/bin/abacus
```

**Windows:**
```powershell
# Extract the archive
tar -xzf abacus_*.tar.gz

# Move to a directory in PATH, e.g.:
Move-Item abacus.exe "$env:LOCALAPPDATA\Programs\"
```

### Method 4: Go Install

If you have Go installed, you can use `go install`:

```bash
go install github.com/ChrisEdwards/abacus/cmd/abacus@latest
```

This command:
- Downloads the source code
- Compiles the binary
- Installs it to `$GOPATH/bin` (typically `~/go/bin`)

**Add to PATH:**

If `abacus` command is not found, add Go's bin directory to your PATH:

**For bash (~/.bashrc or ~/.bash_profile):**
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

**For zsh (~/.zshrc):**
```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

**For fish (~/.config/fish/config.fish):**
```fish
set -gx PATH $PATH (go env GOPATH)/bin
```

After editing, reload your shell configuration:
```bash
source ~/.bashrc  # or ~/.zshrc, etc.
```

### Method 5: Build from Source

For the latest development version or to contribute (requires Go 1.25.3+):

```bash
# Clone the repository
git clone https://github.com/ChrisEdwards/abacus.git
cd abacus

# Build the binary with version info
make build

# Verify the build
./abacus --version
```

**Install system-wide:**

**Linux and macOS:**
```bash
sudo mv abacus /usr/local/bin/
```

**Alternative: Install to user directory:**
```bash
mkdir -p ~/.local/bin
mv abacus ~/.local/bin/
# Add ~/.local/bin to PATH if not already there
```

**Install using Make:**
```bash
make install  # Installs to $GOPATH/bin
```

### Method 6: Using Make

If you prefer using Make (requires Go 1.25.3+):

```bash
git clone https://github.com/ChrisEdwards/abacus.git
cd abacus

# Build with version injection
make build

# Install to GOPATH/bin
make install

# Or see all available targets
make help
```

**Available make targets:**
- `make build` - Compile the binary to `./abacus`
- `make install` - Install to `$GOPATH/bin`
- `make test` - Run tests
- `make lint` - Run linter
- `make clean` - Remove build artifacts

## Platform-Specific Notes

### macOS

**Recommended:** Use Homebrew for the easiest installation:
```bash
brew tap ChrisEdwards/tap
brew install abacus
```

If you downloaded the binary directly, you may need to allow it to run:

```bash
# If you get a security warning
xattr -d com.apple.quarantine /usr/local/bin/abacus
```

Or open System Preferences → Security & Privacy and allow it.

### Linux

Most Linux distributions work out of the box. For older distributions, you may need to update Go:

```bash
# Ubuntu/Debian
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt install golang-go

# Fedora
sudo dnf install golang

# Arch Linux
sudo pacman -S go
```

### Windows

Abacus works on Windows with appropriate terminal emulators:

**Recommended: Use the install script (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/ChrisEdwards/abacus/main/install.ps1 | iex
```

**Alternative: Manual installation:**
1. Download the Windows binary from [GitHub Releases](https://github.com/ChrisEdwards/abacus/releases/latest)
2. Extract `abacus.exe` to a directory in your PATH
3. Or use Go: `go install github.com/ChrisEdwards/abacus/cmd/abacus@latest`

**Recommended Windows Terminals:**
- Windows Terminal (best support for colors and Unicode)
- ConEmu
- Alacritty

## Verify Installation

After installation, verify Abacus is working:

```bash
# Check version
abacus --version

# View help
abacus --help

# Try running in a Beads project directory
cd /path/to/beads/project
abacus
```

## Updating Abacus

### If installed via Homebrew:

```bash
brew upgrade abacus
```

### If installed via install script:

Re-run the install script:
```bash
curl -fsSL https://raw.githubusercontent.com/ChrisEdwards/abacus/main/scripts/install.sh | bash
```

### If installed via go install:

```bash
go install github.com/ChrisEdwards/abacus/cmd/abacus@latest
```

### If built from source:

```bash
cd /path/to/abacus/source
git pull origin main
make build
sudo mv abacus /usr/local/bin/
```

## Uninstalling

### If installed via Homebrew:

```bash
brew uninstall abacus
```

### If installed via install script:

```bash
rm ~/.local/bin/abacus
# Or if installed elsewhere:
rm /usr/local/bin/abacus
```

### If installed via go install:

```bash
rm $(go env GOPATH)/bin/abacus
```

### Remove configuration (optional):

```bash
rm -rf ~/.config/abacus
```

## Troubleshooting Installation

### "go: command not found"

Go is not installed or not in your PATH. Install Go from [golang.org/dl](https://golang.org/dl).

### "cannot find package"

Your GOPATH might not be set correctly:

```bash
# Check GOPATH
go env GOPATH

# It should output something like /home/username/go
```

### Permission Denied

If you get permission errors during installation:

```bash
# Don't use sudo with go install
# Instead, ensure your GOPATH/bin is writable
mkdir -p $(go env GOPATH)/bin
chmod 755 $(go env GOPATH)/bin
```

### Build Fails

Make sure you have the correct Go version:

```bash
go version
# Should be 1.25.3 or later
```

If needed, update Go from [golang.org/dl](https://golang.org/dl).

## Next Steps

Now that Abacus is installed:

- Read the [Getting Started](getting-started.md) guide
- Configure Abacus with the [Configuration](configuration.md) guide
- Learn all features in the [User Guide](user-guide.md)
