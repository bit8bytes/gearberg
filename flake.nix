{
  description = "gearberg: Self-hostable inventory management.";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
  inputs.nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
  outputs = {
    self,
    nixpkgs,
    nixpkgs-unstable,
  }: let
    supportedSystems = [
      "x86_64-linux"
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
  };
}