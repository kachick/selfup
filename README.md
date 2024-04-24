# selfup

[![CI - Go Status](https://github.com/kachick/selfup/actions/workflows/ci-go.yml/badge.svg?branch=main)](https://github.com/kachick/selfup/actions/workflows/ci-go.yml?query=branch%3Amain+)
[![CI - Nix Status](https://github.com/kachick/selfup/actions/workflows/ci-nix.yml/badge.svg?branch=main)](https://github.com/kachick/selfup/actions/workflows/ci-nix.yml?query=branch%3Amain+)

Replace strings if the line contains the definition of how to update itself.

## Installation

In [Nix](https://nixos.org/) [Flake](https://nixos.wiki/wiki/Flakes), you can skip installation steps

```console
> nix run github:kachick/selfup/v1.1.1 -- --version
selfup dev (rev)
```

Prebuilt binaries are available for download from [releases](https://github.com/kachick/selfup/releases)

```console
> install_path="$(mktemp -d)"
> curl -L https://github.com/kachick/selfup/releases/download/v1.1.1/selfup_Linux_x86_64.tar.gz | tar xvz -C "$install_path" selfup
> "${install_path}/selfup" --version
selfup 1.0.0 (REV)
```

## Usage

```plaintext
selfup [SUB] [OPTIONS] [PATH]...
```

Assume some GitHub actions workflow have lines like follows

```yaml
- uses: dprint/check@v2.2
  with:
    dprint-version: '0.40.2' # selfup { "extract": "\\d[^']+", "replacer": ["dprint", "--version"], "nth": 2 }
```

Then you can call selfup as this

```bash
selfup run .github/workflows/*.yml
```

You can check the running plans with `list` subcommand

```console
> selfup list .github/workflows/*.yml
  .github/workflows/lint.yml:17: 0.40.2
✓ .github/workflows/release.yml:37: 1.20.0 => 1.42.9
  .github/workflows/release.yml:50: 3.3.1

1/3 items will be replaced
```

### JSON schema

| Field     | Type     | Description                                                                                                    |
| --------- | -------- | -------------------------------------------------------------------------------------------------------------- |
| extract   | string   | Golang regex like [RE2](https://github.com/google/re2/wiki/Syntax), remember to escape meta characters in JSON |
| replacer  | []string | Command and the arguments. Use `["bash", "-c", "your_script \| as_using_pipe"]` for script style               |
| nth       | number   | Cut the fields, First is `1`, will work no fields mode by default(`0`)                                         |
| delimiter | string   | Split the STDOUT to make fields, using [strings.Fields](https://pkg.go.dev/strings#Fields) by default(`""`)    |

### Options

- `--prefix`: Set customized prefix to begin the JSON
- `--skip-by`: Skips to parse JSON and runs if the line includes this string
- `--check`: Exit with non 0 value if found changes or the plans
- `--no-color`: Avoid to wrap colors even if executed in terminal
- `--version`: Print the version

## Examples

- [examples](examples)
- [workflow file with v0.0.2](https://github.com/kachick/anylang-template/blob/0d50545d31a5b7b878d2738db38654c23cd37ef4/.github/workflows/reusable-update-nixpkgs-and-versions-in-ci.yml#L68) => [result PR](https://github.com/kachick/anylang-template/pull/24)

## FAQ

- `selfup run .github` does not work. Is there walker option?
  - Just taking target paths, recommend to use as this `git ls-files .github | xargs selfup run`

- What are the advantages over version updaters?
  - [dependabot does not have this feature](https://github.com/dependabot/dependabot-core/issues/9557)
  - [renovatebot only has it in self-hosted runners](https://github.com/renovatebot/renovate/issues/5004)
  - In my use case, I need to sync the versions with nixpkgs, not sync with the latest.\
    Both renovatebot and dependabot will not fit for this use.

## Motivation

I'm using this tool to update tool versions in several GitHub actions.\
Especially I want to synchronize them with Nix shells.

Nix and the ecosystem provide useful CIs, but the runtime footprint is not small even for small changes.\
So I'm currently using both Nix CI and some tools CIs.
