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
