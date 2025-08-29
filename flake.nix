{
  description = "Readn - RSS reader with AI integration and discussion support";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        packages.default = pkgs.buildGoModule rec {
          pname = "readn";
          version = "dev";

          src = ./.;

          vendorHash = null;

          ldflags = [
            "-s"
            "-w"
            "-X main.Version=${version}"
            "-X main.GitHash=none"
          ];

          tags = [
            "sqlite_foreign_keys"
            "sqlite_json"
          ];

          meta = with pkgs.lib; {
            description = "RSS reader with AI integration and discussion support";
            mainProgram = "readn";
            homepage = "https://github.com/thang-qt/Readn";
            license = licenses.mit;
          };
        };

        apps.default = flake-utils.lib.mkApp {
          drv = self.packages.${system}.default;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.gopls
            pkgs.git
          ];
        };
      });
}