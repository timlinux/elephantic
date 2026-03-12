{
  description = "Elephantic - MCP server for counting points by country";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            geos
            sqlite
            pkg-config
            gcc
          ];

          shellHook = ''
            export CGO_ENABLED=1
            export CGO_CFLAGS="$(pkg-config --cflags geos)"
            export CGO_LDFLAGS="$(pkg-config --libs geos)"
            echo "Elephantic development environment ready"
            echo "Run 'go build' to compile"
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "elephantic";
          version = "1.0.0";
          src = ./.;
          vendorHash = null; # Will need to be updated after go mod tidy

          nativeBuildInputs = with pkgs; [ pkg-config ];
          buildInputs = with pkgs; [ geos sqlite ];

          CGO_ENABLED = 1;

          meta = with pkgs.lib; {
            description = "MCP server for counting GeoJSON points by country";
            homepage = "https://github.com/kartoza/elephantic";
            license = licenses.mit;
            maintainers = [ ];
          };
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/elephantic";
        };
      }
    );
}
