{
  lib,
  buildGo124Module,
  versionCheckHook,
}:

let
  mainProgram = "selfup";
in
buildGo124Module (finalAttrs: {
  pname = "selfup";
  version = "1.2.0";
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
    "-s"
    "-w"
    "-X main.version=v${finalAttrs.version}"
  ];

  # When updating go.mod or go.sum, update this sha together with `nix-update selfup --version=skip --flake`
  vendorHash = "sha256-fkdmmjQJE51irYRT/uY83kgOd9GjU9DQWiYGnP18Iqk=";

  # https://github.com/kachick/times_kachick/issues/316
  env.CGO_ENABLED = "0";

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
})
