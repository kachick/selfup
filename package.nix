{
  lib,
  buildGo123Module,
  versionCheckHook,
}:

let
  mainProgram = "selfup";
in
# TODO: Replace with go 1.24(GH-327) and finalAttrs
buildGo123Module rec {
  pname = "selfup";
  version = "1.1.9";
  src = lib.fileset.toSource {
    root = ./.;
    # - Don't just use `fileset.gitTracked root`, then always rebuild even if just changed the README.md
    # - Don't use gitTracked for now, even if filtering with intersection, the feature is not supported in nix-update. See https://github.com/Mic92/nix-update/issues/335
    fileset = lib.fileset.unions [
      ./go.mod
      ./go.sum
      ./cmd
      ./internal
    ];
  };
  # src = lib.cleanSource self; # Requires this old style if I use nix-update
  ldflags = [
    "-X main.version=v${version}"
    "-X main.commit=${"0000000000000000000000000000000000000000"}" # TODO: Remove these revision in version format
  ];

  # When updating go.mod or go.sum, update this sha together with `nix-update selfup --version=skip --flake`
  vendorHash = "sha256-rLS2bLpPM0Uo/fhLXTwBTimO0r8Y3IvYvMa3mK36DyQ=";

  # https://github.com/kachick/times_kachick/issues/316
  # TODO: Use env after nixos-25.05. See https://github.com/NixOS/nixpkgs/commit/905dc8d978b38b0439905cb5cd1faf79163e1f14#diff-b07c2e878ff713081760cd5dcf0b53bb98ee59515a22e6007cc3d974e404b220R24
  CGO_ENABLED = 0;

  nativeInstallCheckInputs = [
    versionCheckHook
  ];
  doInstallCheck = true;
  versionCheckProgram = "${placeholder "out"}/bin/${mainProgram}";
  versionCheckProgramArg = "--version";

  meta = {
    inherit mainProgram;
    description = "CLI to bump versions";
    homepage = "https://github.com/kachick/selfup";
    license = lib.licenses.mit;
  };
}
