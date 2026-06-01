{
  description = "A flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };
        ldPath =
          with pkgs;
          lib.makeLibraryPath [
            stdenv.cc.cc
            zlib
            glib
            libxcb
            libglvnd
          ];
      in
      {
        devShells.default = pkgs.mkShell {
          LD_LIBRARY_PATH = ldPath;

          packages = with pkgs; [
            nixd

            go
            gopls
            gotools
            golangci-lint
            protobuf
            buf

            bun

            just
          ];

          buildInputs = [ pkgs.bashInteractive ];

          env = {
            GOPROXY = "https://proxy.golang.org,direct";
            GOSUMDB = "sum.golang.org";
          };

          shellHook = ''
            export GOPATH="$PWD/.go"
            export GOBIN="$GOPATH/bin"

            mkdir -p "$GOBIN"
            export PATH="$GOBIN:$PATH"
          '';
        };
      }
    );
}
