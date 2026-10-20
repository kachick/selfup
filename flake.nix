{
  nixConfig = {
    extra-substituters = [
      "https://kachick.cachix.org"
    ];
    extra-trusted-public-keys = [
      "kachick.cachix.org-1:7giZGKW03gPFkbde09JNpDKCbDGGwBuU/Z11W/jUtEQ="
    ];
  };

  inputs = {
    # How to update the revision
    #   - `nix flake update --commit-lock-file` # https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-flake-update.html
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.zst";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      forAllSystems = nixpkgs.lib.genAttrs nixpkgs.lib.systems.flakeExposed;
    in
    rec {
      formatter = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        pkgs.writeShellScriptBin "dprint-fmt" ''
          exec "${nixpkgs.lib.getExe pkgs.dprint}" fmt "$@"
        ''
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default =
            with pkgs;
            mkShell {
              env = {
                # Fix nixd pkgs versions in the inlay hints
                NIX_PATH = "nixpkgs=${pkgs.path}";

                # For vscode typos extension
                TYPOS_LSP_PATH = pkgs.lib.getExe pkgs.typos-lsp;

                # For local development, only defined in package.nix is not useful for daily development
                CGO_ENABLED = "0";
              };

              inputsFrom = [
                self.packages.${system}.selfup
              ];

              buildInputs = [
                # https://github.com/NixOS/nix/issues/730#issuecomment-162323824
                bashInteractive
                nixd
                nix-update
                cachix

                dprint
                goreleaser
                typos
                shellcheck
                go-task
                zizmor
                pinact
              ];
            };
        }
      );

      overlays = {
        default = final: _prev: {
          selfup = final.callPackage ./package.nix { };
        };
      };

      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        rec {
          selfup = pkgs.callPackage ./package.nix { };
          default = selfup;

          update-deps = pkgs.writeShellApplication {
            name = "update-deps";
            runtimeInputs = [
              selfup
              pkgs.git
              pkgs.findutils
              pkgs.dprint
              pkgs.typos
              pkgs.goreleaser
            ];
            text = ''
              git ls-files -z .github | xargs --null selfup run "$@" --
            '';
          };
        }
      );

      # `nix run`
      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = nixpkgs.lib.getExe packages.${system}.selfup;
        };
        update-deps = {
          type = "app";
          program = nixpkgs.lib.getExe packages.${system}.update-deps;
        };
      });
    };
}
