{
  description = "Yarr - Yet another RSS reader (Go)";

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
          pname = "yarr";
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
            description = "Yet another RSS reader";
            mainProgram = "yarr";
            homepage = "https://github.com/nkanaev/yarr";
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