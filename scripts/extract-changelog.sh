#!/usr/bin/env bash
set -euo pipefail

# Extract a specific version section from CHANGELOG.md
# Usage: ./extract-changelog.sh <version>
# Example: ./extract-changelog.sh 0.2.0

if [ $# -eq 0 ]; then
    echo "Error: Version argument required" >&2
    echo "Usage: $0 <version>" >&2
    echo "Example: $0 0.2.0" >&2
    exit 1
fi

VERSION="$1"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CHANGELOG="$SCRIPT_DIR/../CHANGELOG.md"

if [ ! -f "$CHANGELOG" ]; then
    echo "Error: CHANGELOG.md not found at $CHANGELOG" >&2
    exit 1
fi

# Extract the section for the specified version
# Start from ## [VERSION] and continue until the next ## [ line (or end of file)
awk -v version="$VERSION" '
    BEGIN { found=0; printing=0 }

    # Match the version header
    /^## \[/ {
        if ($0 ~ "\\[" version "\\]") {
            found=1
            printing=1
            next  # Skip the header itself
        } else if (printing) {
            # Hit the next version section, stop
            exit
        }
    }

    # Stop at link reference section
    /^\[.*\]:[ ]*https?:/ {
        if (printing) {
            exit
        }
    }

    # Print lines while in the target section
    printing { print }

    END {
        if (!found) {
            print "Error: Version [" version "] not found in CHANGELOG.md" > "/dev/stderr"
            exit 1
        }
    }
' "$CHANGELOG"
