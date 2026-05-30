let
  pkgs = import <nixpkgs> {};
  rev = builtins.fetchGit { url = ./.; }.shortRev;
in import ./nixos/web.nix {
  buildpkgs = pkgs;
  version = rev;
}
