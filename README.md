# kubectl-finalize

A kubectl plugin to force delete Kubernetes resources that are stuck in a Terminating state. This plugin is particularly useful when dealing with namespaces or resources that won't delete due to stuck finalizers.

## Features

- Force deletes resources stuck in Terminating state
- Safely removes finalizers from resources
- Validates namespace state before attempting deletion
- Works with any Kubernetes resource type
- Supports custom resources via API discovery

## Installation

### Using Go

```bash
go install github.com/yourusername/kubectl-finalize@latest
```

### Manual Installation

1. Download the latest release for your platform from the [releases page](https://github.com/yourusername/kubectl-finalize/releases)
2. Make it executable: `chmod +x kubectl-finalize`
3. Move it to your PATH: `sudo mv kubectl-finalize /usr/local/bin/`

## Usage

```bash
# Force delete a stuck namespace
kubectl finalize namespace/stuck-namespace

# Force delete a pod in a specific namespace
kubectl finalize pod/stuck-pod -n my-namespace

# Force delete a deployment
kubectl finalize deployment/stuck-deployment -n my-namespace
```

### Safety Features

The plugin includes several safety checks:
- For namespaces, it verifies that:
  - The namespace exists
  - The namespace is in "Terminating" state
  - The namespace has a deletion timestamp
- For other resources:
  - Validates resource existence
  - Removes finalizers safely
  - Uses background propagation policy

## Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/kubectl-finalize
cd kubectl-finalize

# Build
just build

# Install
just install
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details 