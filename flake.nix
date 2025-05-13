{
  inputs = {
    # How to update the revision
    #   - `nix flake update --commit-lock-file` # https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-flake-update.html
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { nixpkgs, ... }:
    let
      forAllSystems = nixpkgs.lib.genAttrs nixpkgs.lib.systems.flakeExposed;
    in
    rec {
      formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.nixfmt-tree);
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
              };
              buildInputs = [
                # https://github.com/NixOS/nix/issues/730#issuecomment-162323824
                bashInteractive
                nixd
                nixfmt-rfc-style
                nix-update

                go_1_24
                dprint
                goreleaser
                typos
                go-task
              ];
            };
        }
      );

      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        rec {
          selfup = pkgs.callPackage ./package.nix { };
          default = selfup;
        }
      );

      # `nix run`
      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = nixpkgs.lib.getExe packages.${system}.selfup;
        };
      });
    };
}
