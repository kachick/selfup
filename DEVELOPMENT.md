# How to develop

## Setup

1. Install [Nix](https://nixos.org/) package manager
2. Run `nix develop` or `direnv allow` in project root
3. You can use development tools

```console
> nix develop
(prepared shell)

> task
task: [build] ..."
task: [test] go test
task: [lint] dprint check
task: [lint] go vet
PASS
ok      selfup    0.313s

> task run -- --version
selfup dev (rev)
```

## Binary cache for released tags

To build and pin binary cache for a released tag into Cachix:

```console
gh workflow run ci-nix.yml -f target-tag=v1.3.1
```

Always run this workflow on the `main` branch. Using `gh` prevents choosing the wrong branch in the GitHub Web UI.

Note:

- Pins use the `selfup-<version>-<system>` format (e.g. `selfup-1.3.1-x86_64-linux`).
- Only official release tags matching `^v[0-9]+\.[0-9]+\.[0-9]+$` are pushed and pinned (pre-releases such as `-beta` are skipped).
