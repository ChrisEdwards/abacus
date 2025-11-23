# Installation Guide

This guide covers various methods for installing Abacus on different platforms.

## Prerequisites

### Required

- **Go 1.25.3 or later** - Required for building and installing
- **Beads CLI v0.24.0 or later** - Install from [github.com/steveyegge/beads](https://github.com/steveyegge/beads) and verify with `bd --version`

### Recommended

- Terminal with 256-color support
- Git (for building from source)

### Verify Prerequisites

```bash
# Check Go version
go version

# Check Beads installation
bd --version

# Check terminal color support
echo $TERM
```

## Installation Methods

### Method 1: Go Install (Recommended)

The easiest way to install Abacus is using Go's built-in install command:

```bash
go install github.com/yourusername/abacus/cmd/abacus@latest
```

This command:
- Downloads the source code
- Compiles the binary
- Installs it to `$GOPATH/bin` (typically `~/go/bin`)

#### Add to PATH

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

### Method 2: Build from Source

For the latest development version or to contribute:

```bash
# Clone the repository
git clone https://github.com/yourusername/abacus.git
cd abacus

# Build the binary
go build -o abacus ./cmd/abacus

# Verify the build
./abacus --help
```

#### Install System-Wide

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

### Method 3: Using Make

If you prefer using Make:

```bash
git clone https://github.com/yourusername/abacus.git
cd abacus

# Build
make build

# Install to GOPATH/bin
make install

# Or see all available targets
make help
```

Available make targets:
- `make build` - Compile the binary to `./abacus`
- `make install` - Install to `$GOPATH/bin`
- `make test` - Run tests
- `make lint` - Run linter
- `make clean` - Remove build artifacts

## Platform-Specific Notes

### macOS

On macOS, you may need to allow the binary to run:

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

While Abacus is primarily designed for Unix-like systems, it can work on Windows with appropriate terminal emulators:

1. **Install Go** from [golang.org/dl](https://golang.org/dl)
2. **Use Windows Terminal** or another modern terminal emulator
3. **Install via Go:**
   ```powershell
   go install github.com/yourusername/abacus/cmd/abacus@latest
   ```
4. Add `%USERPROFILE%\go\bin` to your PATH

**Recommended Windows Terminals:**
- Windows Terminal (best support for colors and Unicode)
- ConEmu
- Alacritty

## Verify Installation

After installation, verify Abacus is working:

```bash
# Check version and help
abacus --help

# Try running in a test directory
cd /path/to/beads/project
abacus
```

## Updating Abacus

### If installed via go install:

```bash
go install github.com/yourusername/abacus/cmd/abacus@latest
```

### If built from source:

```bash
cd /path/to/abacus/source
git pull
go build -o abacus ./cmd/abacus
sudo mv abacus /usr/local/bin/
```

## Uninstalling

### If installed via go install:

```bash
rm $(go env GOPATH)/bin/abacus
```

### If installed system-wide:

```bash
sudo rm /usr/local/bin/abacus
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
