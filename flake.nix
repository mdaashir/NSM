{
  description = "NSM (Nix Shell Manager) - A tool to manage Nix development environments";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "nsm";
          version = "1.1.6";
          src = ./.;

          vendorHash = null;

          ldflags = ["-s" "-w"];

          meta = with pkgs.lib; {
            description = "A tool to manage Nix development environments";
            homepage = "https://github.com/mdaashir/NSM";
            license = licenses.mit;
            maintainers = with maintainers; [ mdaashir ];
            platforms = platforms.all;
          };
        };

        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            golangci-lint
            goreleaser
          ];
        };
      }
    );
}
