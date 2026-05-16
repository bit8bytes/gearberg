# This module defines the available packages in the dev shell.
# Use `nix develop` to open the dev shell.
{pkgs, ...}:
pkgs.mkShellNoCC {
  packages = with pkgs; [
    go
    gcc
    govulncheck
    gosec
    gopls
    sqlite
    goose
    sqlc
    alejandra
    tailwindcss_4
    reflex
    watchman
    golangci-lint
    go-licenses
  ];
  shellHook = ''
    echo "Welcome to the dev shell! All required tools are available."
  '';
}
