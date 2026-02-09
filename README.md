# selfup

[![CI - Go Status](https://github.com/kachick/selfup/actions/workflows/ci-go.yml/badge.svg?branch=main)](https://github.com/kachick/selfup/actions/workflows/ci-go.yml?query=branch%3Amain+)
[![CI - Nix Status](https://github.com/kachick/selfup/actions/workflows/ci-nix.yml/badge.svg?branch=main)](https://github.com/kachick/selfup/actions/workflows/ci-nix.yml?query=branch%3Amain+)
[![Release](https://github.com/kachick/selfup/actions/workflows/release.yml/badge.svg)](https://github.com/kachick/selfup/actions/workflows/release.yml)

Replace strings in files using update rules defined in comments.

## Installation

### Nix

In [Nix](https://nixos.org/) [Flake](https://nixos.wiki/wiki/Flakes), you can skip installation steps

```console
> nix run github:kachick/selfup/v1.3.0 -- --version
selfup v1.3.0
```

You can also use the binary cache defined in the [flake](flake.nix).\
This is available for recent tagged versions, but only if the user is a trusted-user in your nix.conf file.

```bash
> grep trusted-users /etc/nix/nix.conf
trusted-users = root your_user

> nix run --accept-flake-config github:kachick/selfup/v1.3.0 -- --version
selfup v1.3.0
```

### Prebuilt-binary

Prebuilt binaries are available for download from [releases](https://github.com/kachick/selfup/releases)

```console
> install_path="$(mktemp -d)"
> curl -L https://github.com/kachick/selfup/releases/download/v1.3.0/selfup_Linux_x86_64.tar.gz | tar xvz -C "$install_path" selfup
> "${install_path}/selfup" --version
selfup VERSION
```

You can also use [gh](https://github.com/cli/cli) command.

```bash
gh release download 'v1.3.0' --pattern 'selfup_Linux_x86_64.tar.gz' --repo kachick/selfup
tar -xvzf 'selfup_Linux_x86_64.tar.gz'
```

## Usage

```plaintext
selfup [SUB] [OPTIONS] [PATH]...
```

Assume a GitHub Actions workflow has lines like this:

```yaml
- uses: dprint/check@v2.2
  with:
    dprint-version: '0.40.2' # selfup { "extract": "\\b[0-9.]+", "replacer": ["dprint", "--version"], "nth": 2 }
```

You can run selfup like this:

```bash
selfup run .github/workflows/*.yml
```

You can check the plans with the `list` subcommand:

```console
> selfup list .github/workflows/*.yml
  .github/workflows/lint.yml:17: 0.40.2
✓ .github/workflows/release.yml:37: 1.20.0 => 1.42.9
  .github/workflows/release.yml:50: 3.3.1

1/3 items will be replaced
```

### JSON schema

| Field     | Type     | Description                                                                                                     |
| --------- | -------- | --------------------------------------------------------------------------------------------------------------- |
| extract   | string   | Golang regex like [RE2](https://github.com/google/re2/wiki/Syntax). Remember to escape meta-characters in JSON. |
| replacer  | []string | Command and arguments. Use `["bash", "-c", "your_script \| as_using_pipe"]` for script style.                   |
| nth       | number   | Field number. The first field is `1`. By default, it uses the whole line (`0`).                                 |
| delimiter | string   | Separator to split STDOUT into fields. It uses [strings.Fields](https://pkg.go.dev/strings#Fields) by default.  |

### Options

- `--prefix`: Set a custom prefix pattern (RE2) before the JSON.
- `--skip-by`: Skip lines that contain this string.
- `--check`: Exit with a non-zero code if changes or plans are found.
- `--no-color`: Disable colored output.
- `--version`: Print the version.

## Examples

- [examples](examples)
- [workflow file with v0.0.2](https://github.com/kachick/anylang-template/blob/0d50545d31a5b7b878d2738db38654c23cd37ef4/.github/workflows/reusable-update-nixpkgs-and-versions-in-ci.yml#L68) => [result PR](https://github.com/kachick/anylang-template/pull/24)

## FAQ

- `selfup run .github` does not work. Is there a walker option?
  - It only takes target paths. One way to use it is: `git ls-files -z .github | xargs --null selfup run --`

- What are the advantages over other version updaters?
  - [Dependabot does not have this feature.](https://github.com/dependabot/dependabot-core/issues/9557)
  - [Renovate only has it in self-hosted runners.](https://github.com/renovatebot/renovate/issues/5004)
  - In my case, I need to sync versions with nixpkgs, not always the latest.\
    Both Renovate and Dependabot do not fit this use case.

## Motivation

I use this tool to update tool versions in several GitHub Actions.
I especially want to synchronize them with Nix shells.

Nix and its ecosystem provide useful CI, but the runtime footprint is not small even for small changes.\
So I currently use both Nix CI and some tool-specific CIs.
