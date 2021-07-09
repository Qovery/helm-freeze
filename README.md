# helm-freeze
[![Go Report Card](https://goreportcard.com/badge/github.com/Qovery/helm-freeze)](https://goreportcard.com/report/github.com/Qovery/helm-freeze)
[![Powered](https://img.shields.io/badge/Powered%20by-Qovery-blueviolet)](https://www.qovery.com)

<p align="center">
    <img src="./helm_freeze_logo.png" width=420 />
</p>

Freeze your charts in the wished versions. Helm freeze helps you to declare the charts
you want to use in a desired version and download them locally. This to freeze/lock them
directly in your Git repository.

The advantages are:
* Follow GitOps philosophy
* Know exactly what has changed between 2 charts version with a `git diff`
* One place to list them all
* Works well with monorepo
* Declarative configuration (YAML file)
* Supports git repositories in addition to charts repositories

## Installation

### Mac OS
On Mac, you need to have [brew](https://brew.sh/) installed, then you can run those commands:
```bash
brew tap Qovery/helm-freeze
brew install helm-freeze
```

### Arch Linux
An AUR package exists called `helm-freeze`, you can install it with `yay`:
```bash
yay helm-freeze
```

### Others
You can download binaries from the [release section](https://github.com/Qovery/helm-freeze/releases).

## Usage
To use helm-freeze, you need a configuration file. You can generate a default file like this:

```shell script
helm-freeze init
```
A minimal file named `helm-freeze.yaml` will be generated. Here is an example of a more complex one:

```yaml
charts:
    # Chart name
  - name: cert-manager
    # Chart version
    version: v0.16.0
    # The repo to use (declared below in the repos section)
    repo_name: jetstack
    # No destinations is declared, the default one will be used
    comment: "You can add comments"
  - name: fluent-bit
    repo_name: lifen
    version: 2.8.0
    # If you temporary want to stop syncing a specific chart
    no_sync: true
  - name: nginx-ingress
    # No repo_name is specified, stable will be used
    version: 1.35.0
    # Change the destination to another one (declared in destinations section)
    dest: custom
  - name: pleco
    repo_name: git-repo
    # When using a git repo, chart_path is mandatory, you need to specify the chart folder path
    chart_path: /charts/pleco
    dest: custom
    version: v0.8.4

repos:
    # Stable is the default one
  - name: stable
    url: https://charts.helm.sh/stable
  - name: jetstack
    url: https://charts.jetstack.io
  - name: lifen
    url: https://honestica.github.io/lifen-charts
  - name: git-repo
    url: https://github.com/Qovery/pleco.git
    # If you want to directly use a chart folder in a git repo, set type to git
    type: git

destinations:
  - name: default
    path: /my/absolute/path
  - name: custom
    path: ./my/relative/path
```

Then use `sync` arg to locally download the declared versions, here is an example:
```bash
$ helm-freeze sync

[+] Adding helm repos
 -> stable
 -> aws

[+] Updating helm repos

[+] Downloading charts
 -> stable/nginx-ingress 1.35.0
 -> stable/prometheus-operator 8.15.12
 -> stable/elasticsearch-curator 2.1.5
 -> aws/aws-node-termination-handler 0.8.0
 -> aws/aws-vpc-cni 1.0.9
 -> git/pleco 0.8.4

Sync succeed!
```

If you update a chart, launch `sync` and you'll be able to see the differences with `git diff`.
