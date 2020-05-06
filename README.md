# kubectl-nse
A convenience `kubectl` plugin to execute `nsenter` against a Pod's container.

## Prerequisites
- SSH login information for the cluster's nodes must be present in `.ssh/config`
- The node user must be able to run either of `docker` or `crictl` (depending on your cluster)
- `nsenter` must be available on the node

## Usage:
```
  kubectl nse [flags]

Flags:
  -c, --container string   specify container name for pods with multiple containers
  -h, --help               help for kubectl
      --kind               the target cluster is a kind cluster
  -n, --namespace string   specify namespace
      --node string        specify node to which the pod belongs
      --nsargs string      quoted options to pass to nsenter
  -l, --selector string    selector to filter on
      --sudo               execute commands on the remote host with sudo
```

## Installation
Use the binary directly or copy it to a location inside `$PATH` to use with `kubectl`. Binary releases will one day be uploaded.