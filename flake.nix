{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
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

        monitorCliPkg = pkgs.buildGoModule {
          pname = "monitor-cli";
          version = "0.1.0";
          src = ./.;

          subPackages = [
            "./cmd/monitor"
            "./cmd/relay"
          ];

          vendorHash = null;
          go = pkgs.go;
        };
      in {
        packages = {
          default = monitorCliPkg;
          monitor = monitorCliPkg;
          relay = monitorCliPkg;
        };

        apps = {
          default = {
            type = "app";
            program = "${monitorCliPkg}/bin/monitor";
          };
          monitor = {
            type = "app";
            program = "${monitorCliPkg}/bin/monitor";
          };
          relay = {
            type = "app";
            program = "${monitorCliPkg}/bin/relay";
          };
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
