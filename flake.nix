{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = import nixpkgs { inherit system; };

        monitorPkg = pkgs.buildGoModule {
          pname = "monitor";
          version = "0.1.0";
          src = ./.;

          subPackages = ["./cmd/monitor"];

          vendorHash = null;
          go = pkgs.go;
        };
      in {
        packages = {
          default = monitorPkg;
          monitor = monitorPkg;
        };

        apps.default = {
          type = "app";
          program = "${monitorPkg}/bin/monitor";
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
          ];
        };
      }
    );
}
