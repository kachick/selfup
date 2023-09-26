# selfup

[![CI - Go Status](https://github.com/kachick/selfup/actions/workflows/ci-go.yml/badge.svg?branch=main)](https://github.com/kachick/selfup/actions/workflows/ci-go.yml?query=branch%3Amain+)
[![CI - Nix Status](https://github.com/kachick/selfup/actions/workflows/ci-nix.yml/badge.svg?branch=main)](https://github.com/kachick/selfup/actions/workflows/ci-nix.yml?query=branch%3Amain+)

Replace strings if the line contains the definition of how to update itself.

## Installation

In [Nix](https://nixos.org/), you can skip installation steps

```console
> nix run github:kachick/selfup/v0.0.2 -- --version
selfup dev (rev)
```

Prebuilt binaries are available for download from [releases](https://github.com/kachick/selfup/releases)

```console
> install_path="$(mktemp -d)"
> curl -L https://github.com/kachick/selfup/releases/download/v0.0.2/selfup_Linux_x86_64.tar.gz | tar xvz -C "$install_path" selfup
> "${install_path}/selfup" --version
selfup 0.0.2 (REV)
```

## Usage

Assume some GitHub actions workflow have lines like follows

```yaml
- uses: dprint/check@v2.2
  with:
    dprint-version: '0.40.2' # selfup { "regex": "\\d[^']+", "script": "dprint --version | cut -d ' ' -f 2" }
```

Then you can call selfup as this

```bash
selfup run --prefix='# selfup ' .github/workflows/*.yml
```

You can check the running plans with `list` subcommand

```console
> selfup list --prefix='# selfup ' .github/workflows/*.yml
.github/workflows/lint.yml:17: 0.40.2 => 0.40.2 # KEEP
.github/workflows/lint.yml:24: 1.16.12 => 1.16.12 # KEEP
.github/workflows/release.yml:37: 1.20.0 => 999 # UPDATE
.github/workflows/ci-go.yml:48: 2023.1.6 => 2023.1.6 # KEEP
```

### JSON schema

| Field  | Description                                                                               |
| ------ | ----------------------------------------------------------------------------------------- |
| regex  | [RE2](https://github.com/google/re2/wiki/Syntax), be careful to escape for JSON stringify |
| script | Bash script                                                                               |

### Options

`--skip-by` option skips to parse JSON and runs if the line includes this string

```console
> selfup list --prefix='# selfup ' --skip-by=dprint .github/workflows/*.yml
.github/workflows/lint.yml:24: 1.16.12 => 1.16.12 # KEEP
.github/workflows/release.yml:37: 1.20.0 => 999 # UPDATE
.github/workflows/ci-go.yml:48: 2023.1.6 => 2023.1.6 # KEEP
```

## Motivation

I'm using this tool to update tool versions in several GitHub actions.\
Especially I want to synchronize them with Nix shells.

Nix and the ecosystem provide useful CIs, but the runtime footprint is not small even for small changes.\
So I'm currently using both Nix CI and some tools CIs.

You can check actual example at [workflow file](.github/workflows/update-nixpkgs.yml), and the [result PR](https://github.com/kachick/selfup/pull/3).
