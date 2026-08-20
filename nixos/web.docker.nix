# Builds a minimal Docker image containing the gearberg web binary.
# The image targets the same platform as the host by default.
# See https://nix.dev/tutorials/nixos/building-and-running-docker-images.html
{
  pkgs ? import <nixpkgs> {},
  # pkgsTarget must be passed explicitly when called from a flake (pure mode
  # forbids <nixpkgs> lookups). Defaults to a Linux nixpkgs for standalone use.
  pkgsTarget ? import <nixpkgs> {},
  version ? "dev",
  # Target image architecture (e.g. "amd64", "arm64"). Defaults to host arch.
  architecture ? null,
}: let
  gearberg = import ./web.nix {
    buildpkgs = pkgsTarget;
    inherit version;
  };
  imageArgs = {
    name = "gearberg";
    tag = version;
    created = "now"; # If not value "now", the image shows date 1970.
    # Required for outbound TLS: OIDC discovery and JWKS calls will fail without it.
    copyToRoot = pkgs.buildEnv {
      name = "image-root";
      paths = [pkgsTarget.cacert];
    };
    config = {
      # The gearberg binary has sane defaults and will start without environment variables but
      # you can override them if you want to change the defaults. Environment variables can be set
      # in the docker run command or in a docker-compose file.
      Entrypoint = ["${gearberg}/bin/gearberg"];
      # Default subcommand. Override by passing a different command to docker run,
      # e.g. `docker run ghcr.io/bit8bytes/gearberg check`.
      # See `gearberg --help` for all available subcommands.
      Cmd = ["serve"];
      # curl is pulled in via the Nix closure this is why there is no need to add it to copyToRoot.
      # Timing fields are in nanoseconds per the OCI config spec.
      Healthcheck = {
        Test = ["CMD" "${pkgsTarget.curl}/bin/curl" "-f" "http://localhost:8080/v1/healthz"];
        Interval = 30 * 1000000000;
        Timeout = 5 * 1000000000;
        Retries = 3;
      };
      # By default, gearberg listens on port 8080 but this can be overridden by -port flag.
      ExposedPorts = {"8080/tcp" = {};};
      # Env vars can be overwritten on runtime.
      # These are needed here to ensure clean non-docker binary starts
      # without configuring flags or env vars.
      Env = [
        "DB_DSN=file:/data/gearberg.db"
        "STORAGE_DSN=file:///data/uploads"
      ];
      # Gearberg does save its data to /data, which is a volume in the image.
      # When starting the image, you should mount a host directory to /data to persist the data.
      Volumes = {"/data" = {};};
    };
  };
in
  pkgs.dockerTools.buildImage (imageArgs
    // (
      if architecture != null
      then {inherit architecture;}
      else {}
    ))
