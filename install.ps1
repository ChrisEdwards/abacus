# Abacus Installation Script for Windows PowerShell
# Usage: irm https://raw.githubusercontent.com/ChrisEdwards/abacus/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "ChrisEdwards/abacus"
$BinaryName = "abacus"
$InstallDir = "$env:LOCALAPPDATA\Programs\abacus"

# Colors for output
function Write-Info($message) {
    Write-Host "ℹ " -ForegroundColor Blue -NoNewline
    Write-Host $message
}

function Write-Success($message) {
    Write-Host "✓ " -ForegroundColor Green -NoNewline
    Write-Host $message
}

function Write-Warning($message) {
    Write-Host "⚠ " -ForegroundColor Yellow -NoNewline
    Write-Host $message
}

function Write-Error-Custom($message) {
    Write-Host "✗ " -ForegroundColor Red -NoNewline
    Write-Host $message
}

# Detect platform
function Get-Platform {
    Write-Info "Detecting platform..."

    $os = "windows"
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

    Write-Info "Detected platform: $os/$arch"
    return @{OS = $os; Arch = $arch}
}

# Get latest version from GitHub API
function Get-LatestVersion {
    Write-Info "Fetching latest version..."

    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        $version = $release.tag_name

        if ([string]::IsNullOrEmpty($version)) {
            throw "Version not found in API response"
        }

        Write-Success "Latest version: $version"
        return $version
    }
    catch {
        Write-Error-Custom "Failed to fetch latest version: $_"
        exit 1
    }
}

# Try go install
function Install-ViaGo {
    Write-Warning "Trying go install..."

    try {
        $goVersion = go version 2>$null
        if ($LASTEXITCODE -ne 0) {
            Write-Error-Custom "Go is not installed"
            return $false
        }

        Write-Info "Found $goVersion"
        Write-Info "Running: go install github.com/$Repo/cmd/$BinaryName@latest"

        go install "github.com/$Repo/cmd/$BinaryName@latest"

        if ($LASTEXITCODE -eq 0) {
            $goPath = go env GOPATH
            $goBin = Join-Path $goPath "bin"
            $binaryPath = Join-Path $goBin "$BinaryName.exe"

            if (Test-Path $binaryPath) {
                Write-Success "Installed via go install to $goBin"

                # Add to PATH if not already there
                $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
                if ($userPath -notlike "*$goBin*") {
                    Write-Info "Adding $goBin to PATH..."
                    [Environment]::SetEnvironmentVariable("Path", "$userPath;$goBin", "User")
                    $env:Path = "$env:Path;$goBin"
                    Write-Success "Added to PATH (restart terminal for changes to take effect)"
                }

                return $true
            }
        }

        return $false
    }
    catch {
        Write-Error-Custom "go install failed: $_"
        return $false
    }
}

# Download and install binary
function Install-Binary {
    param(
        [string]$Version,
        [hashtable]$Platform
    )

    $versionNumber = $Version -replace '^v', ''
    $archiveName = "${BinaryName}_${versionNumber}_$($Platform.OS)_$($Platform.Arch).tar.gz"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$archiveName"

    Write-Info "Downloading $archiveName..."

    $tempDir = Join-Path $env:TEMP "abacus-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        $archivePath = Join-Path $tempDir $archiveName

        # Download
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -ErrorAction Stop
        Write-Success "Downloaded $archiveName"

        Write-Info "Extracting binary..."

        # Extract using tar (available in Windows 10+)
        Push-Location $tempDir
        try {
            tar -xzf $archiveName 2>$null
            if ($LASTEXITCODE -ne 0) {
                throw "tar extraction failed"
            }
        }
        finally {
            Pop-Location
        }

        $binaryPath = Join-Path $tempDir "$BinaryName.exe"
        if (-not (Test-Path $binaryPath)) {
            # Try without .exe extension (the archive contains just "abacus")
            $binaryPath = Join-Path $tempDir $BinaryName
            if (-not (Test-Path $binaryPath)) {
                throw "Binary not found in archive"
            }
        }

        Write-Success "Extracted binary"

        # Create install directory
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }

        # Backup existing installation
        $targetPath = Join-Path $InstallDir "$BinaryName.exe"
        if (Test-Path $targetPath) {
            Write-Warning "Existing installation found, backing up..."
            Move-Item $targetPath "$targetPath.backup" -Force
        }

        # Install binary
        Write-Info "Installing to $InstallDir..."
        Copy-Item $binaryPath $targetPath -Force

        Write-Success "Installed $BinaryName to $InstallDir"

        # Add to PATH if not already there
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$InstallDir*") {
            Write-Info "Adding $InstallDir to PATH..."
            [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
            $env:Path = "$env:Path;$InstallDir"
            Write-Success "Added to PATH (restart terminal for changes to take effect)"
        }

        return $true
    }
    catch {
        Write-Error-Custom "Installation failed: $_"
        return $false
    }
    finally {
        # Cleanup
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Verify installation
function Test-Installation {
    Write-Info "Verifying installation..."

    $command = Get-Command $BinaryName -ErrorAction SilentlyContinue
    if (-not $command) {
        Write-Warning "$BinaryName is not in PATH yet"
        Write-Host ""
        Write-Host "Please restart your terminal or run:"
        Write-Host "  `$env:Path = [System.Environment]::GetEnvironmentVariable('Path', 'User')" -ForegroundColor Blue
        Write-Host ""
        Write-Host "Or run directly: $InstallDir\$BinaryName.exe"
        return
    }

    try {
        $versionOutput = & $BinaryName --version 2>&1
        Write-Success "$BinaryName is installed and in PATH"
        Write-Host ""
        Write-Host $versionOutput
    }
    catch {
        Write-Warning "Could not verify installation: $_"
    }
}

# Check for existing installation
function Test-ExistingInstallation {
    $command = Get-Command $BinaryName -ErrorAction SilentlyContinue
    if ($command) {
        Write-Warning "Found existing installation:"
        Write-Host "  Location: $($command.Source)"

        try {
            $version = & $BinaryName --version 2>&1 | Select-Object -First 1
            Write-Host "  Version:  $version"
        }
        catch {
            Write-Host "  Version:  unknown"
        }

        Write-Host ""
        $response = Read-Host "Continue with installation? [y/N]"
        if ($response -notmatch '^[Yy]$') {
            Write-Host "Installation cancelled"
            exit 0
        }
    }
}

# Main installation flow
function Main {
    Write-Host ""
    Write-Host "═══════════════════════════════════════"
    Write-Host "  Abacus Installation Script"
    Write-Host "═══════════════════════════════════════"
    Write-Host ""

    Test-ExistingInstallation

    $platform = Get-Platform
    $version = Get-LatestVersion

    Write-Host ""

    # Try direct binary installation first
    if (Install-Binary -Version $version -Platform $platform) {
        Write-Host ""
        Test-Installation
        Write-Host ""
        Write-Success "Installation complete!"
    }
    else {
        # Fall back to go install
        Write-Host ""
        Write-Warning "Direct installation failed"

        if (Install-ViaGo) {
            Write-Host ""
            Test-Installation
            Write-Host ""
            Write-Success "Installation complete via go install!"
        }
        else {
            Write-Host ""
            Write-Error-Custom "Installation failed"
            Write-Host ""
            Write-Host "Please try one of these alternatives:"
            Write-Host "  1. Install Go and run: go install github.com/$Repo/cmd/$BinaryName@latest"
            Write-Host "  2. Download manually from: https://github.com/$Repo/releases"
            Write-Host "  3. Install via Scoop (if available): scoop install abacus"
            exit 1
        }
    }

    Write-Host ""
    Write-Host "For more information:"
    Write-Host "  Documentation: https://github.com/$Repo"
    Write-Host "  Report issues: https://github.com/$Repo/issues"
    Write-Host ""
}

# Run main function
Main
