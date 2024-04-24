{
  inputs = {
    # Candidate channels
    #   - https://github.com/kachick/anylang-template/issues/17
    #   - https://discourse.nixos.org/t/differences-between-nix-channels/13998
    # How to update the revision
    #   - `nix flake update --commit-lock-file` # https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-flake-update.html
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        updaterVersion =
          if (self ? shortRev)
          then self.shortRev
          else "dev";
      in
      rec {
        devShells.default = with pkgs;
          mkShell {
            buildInputs = [
              # https://github.com/NixOS/nix/issues/730#issuecomment-162323824
              bashInteractive

              go_1_22
              nil
              nixpkgs-fmt
              dprint
              actionlint
              go-task
              goreleaser
              typos
            ];
          };

        packages.selfup = pkgs.buildGo122Module {
          pname = "selfup";
          version = updaterVersion;
          src = pkgs.lib.cleanSource self;

          # When updating go.mod or go.sum, update this sha together as following
          # vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          # `pkgs.lib.fakeSha256` returns invalid string in thesedays... :<;
          vendorHash = "sha256-HWedQeslWH3/F+7su/7+we16SGoBnXpbJFEjw1uxI6w=";
        };

        packages.default = packages.selfup;

        # `nix run`
        apps.default = {
          type = "app";
          program = "${packages.selfup}/bin/selfup";
        };
      }
    );
}
