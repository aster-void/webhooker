{
  description = "Simple webhook receiver";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (pkgs: {
        webhooker = import ./packages/webhooker.nix { inherit pkgs; };
        default = self.packages.${pkgs.system}.webhooker;
      });

      devShells = forAllSystems (pkgs: {
        default = import ./devshell.nix { inherit pkgs; };
      });

      nixosModules.webhooker = import ./modules/nixos/webhooker.nix;

      checks = nixpkgs.lib.genAttrs [ "x86_64-linux" "aarch64-linux" ] (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in {
          integration = import ./checks/integration.nix { inherit pkgs; };
        }
      );
    };
}
