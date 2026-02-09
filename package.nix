{
  lib,
  buildGo125Module,
  versionCheckHook,
  self,
}:

let
  mainProgram = "selfup";
in
buildGo125Module (finalAttrs: {
  pname = "selfup";
  version =
    let
      # https://github.com/NixOS/nix/issues/4682#issuecomment-3263194000
      gitRev = toString (self.shortRev or self.dirtyShortRev or self.lastModified or "DEVELOPMENT");
    in
    lib.removePrefix "v" gitRev;
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
    "-X main.version=${finalAttrs.version}"
  ];

  # When updating go.mod or go.sum, update this sha together with `nix-update selfup --version=skip --flake`
  vendorHash = "sha256-Zp8XxBEDBC7RTbF95xov1WNOmOPRPmq/mCrsM/GFQtE=";

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
