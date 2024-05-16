{
  inputs = {
    # Candidate channels
    #   - https://github.com/kachick/anylang-template/issues/17
    #   - https://discourse.nixos.org/t/differences-between-nix-channels/13998
    # How to update the revision
    #   - `nix flake update --commit-lock-file` # https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-flake-update.html
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    edge-nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    {
      self,
      nixpkgs,
      edge-nixpkgs,
    }:
    let
      # Candidates: https://github.com/NixOS/nixpkgs/blob/release-23.11/lib/systems/flake-systems.nix
      forAllSystems = nixpkgs.lib.genAttrs [
        "x86_64-linux"
        "aarch64-linux"
        "i686-linux"
        "x86_64-darwin"
        # I don't have M1+ mac, providing this for macos-14 free runner https://github.com/actions/runner-images/issues/9741
        "aarch64-darwin"
      ];
    in
    rec {
      formatter = forAllSystems (system: edge-nixpkgs.legacyPackages.${system}.nixfmt-rfc-style);
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          edge-pkgs = edge-nixpkgs.legacyPackages.${system};
        in
        {
          default =
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
        }
      );

      packages = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          edge-pkgs = edge-nixpkgs.legacyPackages.${system};
          version = "v1.1.2";
        in
        rec {
          selfup = edge-pkgs.buildGo122Module {
            pname = "selfup";
            src = pkgs.lib.cleanSource self;
            version = version;
            ldflags = [
              "-X main.version=${version}"
              "-X main.commit=${if (self ? rev) then self.rev else "0000000000000000000000000000000000000000"}"
            ];

            # When updating go.mod or go.sum, update this sha together as following
            vendorHash = "sha256-wZb6U2CSYZvFyT7rT1FWMZpKpuAXjINZrfV0breLguI=";
          };

          default = selfup;
        }
      );

      # `nix run`
      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${packages.${system}.selfup}/bin/selfup";
        };
      });
    };
}
