# kubectl-finalize v0.1.0

Initial release of kubectl-finalize, a kubectl plugin to force delete Kubernetes resources that are stuck in a Terminating state.

## Features

- Force deletes resources stuck in Terminating state
- Safely removes finalizers from resources
- Validates namespace state before attempting deletion
- Works with any Kubernetes resource type
- Supports custom resources via API discovery

## Installation

1. Download the appropriate binary for your platform
2. Make it executable: `chmod +x kubectl-finalize-*`
3. Move it to your PATH: `sudo mv kubectl-finalize-* /usr/local/bin/kubectl-finalize`

## Checksums

```
0363fad51d615557902e5d735d09a523539b9cba1897fefcca6298f6eb983d49  kubectl-finalize-darwin-amd64
454964eee7a50ef3c125b272fc8da2b74dcdd2a1438e7bbf5229b0d898e27b7b  kubectl-finalize-darwin-arm64
773bc2b86b4b621738789563c27ac20d4ba765d19d9ec623cf480c4c6783344f  kubectl-finalize-linux-amd64
1b547128e08caf4468fbbd63e3392bf81ddbafeba9cace602bc5dea83ed26b05  kubectl-finalize-windows-amd64.exe
```

## Usage

```bash
# Force delete a stuck namespace
kubectl finalize namespace/stuck-namespace

# Force delete a pod in a specific namespace
kubectl finalize pod/stuck-pod -n my-namespace

# Force delete a deployment
kubectl finalize deployment/stuck-deployment -n my-namespace
``` 