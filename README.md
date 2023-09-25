# selfup

[![CI - Go Status](https://github.com/kachick/selfup/actions/workflows/ci-go.yml/badge.svg?branch=main)](https://github.com/kachick/selfup/actions/workflows/ci-go.yml?query=branch%3Amain+)
[![CI - Nix Status](https://github.com/kachick/selfup/actions/workflows/ci-nix.yml/badge.svg?branch=main)](https://github.com/kachick/selfup/actions/workflows/ci-nix.yml?query=branch%3Amain+)

Replace strings if the line contains the definition of how to update itself.

## Installation

In [Nix](https://nixos.org/), you can skip installation steps

```bash
nix run github:kachick/selfup -- --prefix='# selfup ' .github/workflows/*.yml
```

## Usage

Assume some GitHub actions workflow have lines like follows

```yaml
- uses: dprint/check@v2.2
  with:
    dprint-version: '0.40.2' # selfup { "regex": "\\d[^']+", "script": "dprint --version | cut -d ' ' -f 2" }
```

Then we can call this tool as follows

```bash
selfup --prefix='# selfup ' .github/workflows/*.yml
```

### JSON schema

| Field  | Description                                                                               |
| ------ | ----------------------------------------------------------------------------------------- |
| regex  | [RE2](https://github.com/google/re2/wiki/Syntax), be careful to escape for JSON stringify |
| script | Bash script                                                                               |

### Mode

`--list-targets` option prints extracted targets without side-effect

```console
> selfup --list-targets --prefix='# selfup ' .github/workflows/*.yml
.github/workflows/ci-go.yml:48: 2023.1.6
.github/workflows/lint.yml:17: 0.40.2
.github/workflows/lint.yml:24: 1.16.11
.github/workflows/release.yml:37: 1.20.0
```

## Motivation

I'm using this tool to update tool versions in several GitHub actions.\
Especially I want to synchronize them with Nix shells.

Nix and the ecosystem provide useful CIs, but the runtime footprint is not small even for small changes.\
So I'm currently using both Nix CI and some tools CIs.
