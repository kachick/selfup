{
  inputs = {
    # Candidate channels
    #   - https://github.com/kachick/anylang-template/issues/17
    #   - https://discourse.nixos.org/t/differences-between-nix-channels/13998
    # How to update the revision
    #   - `nix flake update --commit-lock-file` # https://nixos.org/manual/nix/stable/command-ref/new-cli/nix3-flake-update.html
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.11";
  };

  outputs =
    { self, nixpkgs }:
    let
      forAllSystems = nixpkgs.lib.genAttrs nixpkgs.lib.systems.flakeExposed;
    in
    rec {
      formatter = forAllSystems (system: nixpkgs.legacyPackages.${system}.nixfmt-rfc-style);
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default =
            with pkgs;
            mkShell {
              buildInputs = [
                # https://github.com/NixOS/nix/issues/730#issuecomment-162323824
                bashInteractive
                nil
                nixfmt-rfc-style
                nix-update

                go_1_23
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
          version = "1.1.9";
        in
        rec {
          selfup = pkgs.buildGo123Module {
            pname = "selfup";
            src = pkgs.lib.cleanSource self;
            version = version;
            ldflags = [
              "-X main.version=v${version}"
              "-X main.commit=${if (self ? rev) then self.rev else "0000000000000000000000000000000000000000"}"
            ];

            # When updating go.mod or go.sum, update this sha together with `nix-update selfup --version=skip --flake`
            vendorHash = "sha256-EEpkBezmwGr09xbFKdzL5ntbrVqirMQnkUUI5yFdWBI=";

            # https://github.com/kachick/times_kachick/issues/316
            # TODO: Use env after nixos-25.05. See https://github.com/NixOS/nixpkgs/commit/905dc8d978b38b0439905cb5cd1faf79163e1f14#diff-b07c2e878ff713081760cd5dcf0b53bb98ee59515a22e6007cc3d974e404b220R24
            CGO_ENABLED = 0;

            meta.mainProgram = "selfup";
          };

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
