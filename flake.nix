{
  description = "🌌 Screensaver";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = { self, nixpkgs, ... }:
    let
      systems = [
        "x86_64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
	  version = "0.1.0";
        in
        {
          default = pkgs.buildGoModule {
            pname = "monolisa";
            inherit version;
            src = self;
            modules = ./gomod2nix.toml;

            ldflags = [
	      "-s"
	      "-w"
            ];

            vendorHash = "sha256-GdV+7ccktqbsDwfNLBU8fEsOjtHXBKiqjn7m1lMFvUU=";

            meta = {
              description = "Screensaver without love :3";
              homepage = "https://github.com/IwnuplyNotTyan/monolisa";
              mainProgram = "hera";
            };
          };
	});
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gotools
              golangci-lint
            ];
          };
        });
    };
}
