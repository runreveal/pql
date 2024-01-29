{
  description = "Pipeline Query Language";

  inputs = {
    nixpkgs.url = "nixpkgs";
    flake-utils.url = "flake-utils";
  };

  outputs = { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        go = pkgs.go_1_21;
      in
      {
        packages.clickhouse = pkgs.clickhouse;

        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.clickhouse
            pkgs.go-tools # staticcheck
            go
            pkgs.gotools  # godoc, etc.
          ];

          hardeningDisable = [ "fortify" ];
        };

        checks.go_test = pkgs.stdenv.mkDerivation {
          name = "pql-go-test";
          src = ./.;
          __impure = true;

          nativeBuildInputs = [
            pkgs.cacert
            go
          ];

          buildPhase = ''
            runHook preBuild

            HOME="$(mktemp -d)"
            go test -mod=readonly -race -v ./...

            runHook postBuild
          '';

          installPhase = ''
            runHook preInstall
            touch "$out"
            runHook postInstall
          '';
        };
      }
    );
}
