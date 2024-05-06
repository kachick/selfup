{
  inputs = {
    # Candidate channels
    #   - https://github.com/kachick/anylang-template/issues/17
    #   - https://discourse.nixos.org/t/differences-between-nix-channels/13998
    # How to update the revision
    #   - `nix flake update --commit-lock-file` # https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-flake-update.html
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    edge-nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      edge-nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        edge-pkgs = edge-nixpkgs.legacyPackages.${system};
      in
      rec {
        formatter = edge-pkgs.nixfmt-rfc-style;
        devShells.default =
          with pkgs;
          mkShell {
            buildInputs = [
              # https://github.com/NixOS/nix/issues/730#issuecomment-162323824
              bashInteractive
              nil
              edge-pkgs.nixfmt-rfc-style

              edge-pkgs.go_1_22
              edge-pkgs.dprint
              edge-pkgs.goreleaser
              edge-pkgs.typos
              go-task
            ];
          };

        packages.selfup = edge-pkgs.buildGo122Module rec {
          pname = "selfup";
          src = pkgs.lib.cleanSource self;
          version = "v1.1.2";
          ldflags = [
            "-X main.version=${version}"
            "-X main.commit=${if (self ? rev) then self.rev else "0000000000000000000000000000000000000000"}"
          ];

          # When updating go.mod or go.sum, update this sha together as following
          # vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
          # `pkgs.lib.fakeSha256` returns invalid string in thesedays... :<;
          vendorHash = "sha256-8LGQB7YUNWpZNj+K/vpiC3N+OQDUsv00qRI9jRJBySE=";
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
