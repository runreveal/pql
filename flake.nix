# Copyright 2024 RunReveal Inc.
# SPDX-License-Identifier: Apache-2.0

{
  description = "Pipeline Query Language";

  inputs = {
    nixpkgs.url = "nixpkgs";
    flake-utils.url = "flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        go = pkgs.go_1_21;
      in
      {
        packages.clickhouse = pkgs.clickhouse;

        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go-tools # staticcheck
            pkgs.gotools  # godoc, etc.
          ];

          inputsFrom = [ self.checks.${system}.go_test ];

          hardeningDisable = [ "fortify" ];
        };

        checks.go_test = pkgs.stdenv.mkDerivation {
          name = "pql-go-test";
          src = ./.;
          __impure = true;

          nativeBuildInputs = [
            pkgs.cacert
            pkgs.clickhouse
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
