{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  packages = with pkgs; [
    go
    gopls
    gnumake
    gcc
    git
    sqlite
  ];

  shellHook = ''
    export CGO_ENABLED=1
  '';
}
