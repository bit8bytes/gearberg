{
  description = "Gearberg: Self-hostable AV equipment and rental software.";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
  inputs.nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = {
    self,
    nixpkgs,
    nixpkgs-unstable,
  }: let
    supportedSystems = [
      "x86_64-linux"
      "aarch64-linux"
      "aarch64-darwin"
    ];
    forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
  in {
    # Local development shell with all required tools.
    # Use `nix develop` to open the dev shell.
    devShells = forAllSystems (system: {
      default = import ./nixos/shells/dev.nix {
        pkgs = nixpkgs-unstable.legacyPackages.${system};
      };
    });

    # Go binaries built with Nix.
    packages = forAllSystems (system: {
      default = import ./nixos/web.nix {
        buildpkgs = nixpkgs-unstable.legacyPackages.${system};
        version = self.rev or self.dirtyRev or "dev";
      };
      www = import ./nixos/www.nix {
        buildpkgs = nixpkgs-unstable.legacyPackages.${system};
        version = self.rev or self.dirtyRev or "dev";
      };
    });
  };
}