# Builds a minimal Docker image containing the gearberg web binary.
# On macOS or non-x86_64-linux hosts you need cross-compilation support.
# See https://nix.dev/tutorials/nixos/building-and-running-docker-images.html
{
  pkgs ? import <nixpkgs> {},
  # pkgsTarget must be passed explicitly when called from a flake (pure mode
  # forbids <nixpkgs> lookups). Defaults to a Linux nixpkgs for standalone use.
  pkgsTarget ? import <nixpkgs> {system = "x86_64-linux";},
  version ? "dev",
}: let
  gearberg = import ./web.nix {
    buildpkgs = pkgsTarget;
    inherit version;
  };
in
  pkgs.dockerTools.buildImage {
    name = "gearberg";
    tag = version;
    config = {
      Cmd = ["${gearberg}/bin/gearberg"];
      ExposedPorts = {"8080/tcp" = {};};
    };
  }
