[![CircleCI](https://circleci.com/gh/Hsn723/kubectl-nse.svg?style=svg&circle-token=25e7f76a28bc51caa424e04161cea2a2b4543d08)](https://circleci.com/gh/Hsn723/kubectl-nse) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/hsn723/kubectl-nse) [![](https://godoc.org/github.com/Hsn723/kubectl-nse?status.svg)](http://godoc.org/github.com/Hsn723/kubectl-nse)
 [![Go Report Card](https://goreportcard.com/badge/github.com/Hsn723/kubectl-nse)](https://goreportcard.com/report/github.com/Hsn723/kubectl-nse)

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
Compile using `make build` or download the binary for your platform from the [releases](https://github.com/Hsn723/kubectl-nse/releases), renaming it accordingly. Then, use the binary directly or copy it to a location inside `$PATH` to use with `kubectl`.
