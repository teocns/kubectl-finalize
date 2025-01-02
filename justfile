# Default recipe to display help information
default:
    @just --list

# Ensure bin directory exists
_ensure-bin:
    mkdir -p bin

# Build the kubectl plugin
build: _ensure-bin
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Building kubectl-finalize..."
    CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/kubectl-finalize ./cmd/kubectl-finalize
    echo "✅ Build successful: bin/kubectl-finalize"

# Install the plugin to /usr/local/bin
install: build
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Installing kubectl-finalize to /usr/local/bin..."
    sudo install -m 755 bin/kubectl-finalize /usr/local/bin/
    echo "✅ Plugin installed successfully! You can now use 'kubectl finalize'"

# Development build and run
dev *ARGS: _ensure-bin
    #!/usr/bin/env bash
    set -euo pipefail
    go run ./cmd/kubectl-finalize {{ARGS}}

# Remove the installed plugin
uninstall:
    @echo "Removing kubectl-finalize from /usr/local/bin..."
    @sudo rm -f /usr/local/bin/kubectl-finalize
    @echo "Plugin uninstalled successfully!"

# Run tests
test:
    go test -v ./...

# Clean build artifacts
clean:
    rm -rf bin/
    go clean

# Format code
fmt:
    go fmt ./...

# Run linter
lint:
    golangci-lint run

# Build for multiple platforms
build-all:
    #!/usr/bin/env bash
    platforms=("linux/amd64" "darwin/amd64" "darwin/arm64" "windows/amd64")
    for platform in "${platforms[@]}"; do
        platform_split=(${platform//\// })
        GOOS=${platform_split[0]}
        GOARCH=${platform_split[1]}
        output_name=bin/kubectl-finalize
        if [ $GOOS = "windows" ]; then
            output_name+='.exe'
        fi
        output_name+="-${GOOS}-${GOARCH}"
        
        echo "Building for $GOOS/$GOARCH..."
        CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o $output_name ./cmd/kubectl-finalize
        echo "✅ Built $output_name"
    done
    echo "✅ All platforms built successfully!"

# Create a new release
release VERSION:
    #!/usr/bin/env bash
    echo "Creating release v{{VERSION}}..."
    git tag -a "v{{VERSION}}" -m "Release v{{VERSION}}"
    git push origin "v{{VERSION}}"
    just build-all

# Setup development environment
setup:
    go mod download
    go mod tidy
    mkdir -p bin

# Run the plugin with arguments (for development)
run *ARGS:
    go run ./cmd/kubectl-finalize {{ARGS}} 