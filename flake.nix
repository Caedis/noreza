{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = inputs: let
    supportedSystems = [
      "x86_64-linux"
    ];
    forEachSupportedSystem = f:
      inputs.nixpkgs.lib.genAttrs supportedSystems (
        system:
          f {
            pkgs = import inputs.nixpkgs {
              inherit system;
            };
          }
      );
  in {
    devShells = forEachSupportedSystem (
      {pkgs}: {
        default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_latest
            gcc
            glibc
            xorg.libX11
            systemd
          ];

          shellHook = ''
            # Enable CGO
            export CGO_ENABLED=1

            # _FORTIFY_SOURCE annoyances
            export CGO_CFLAGS="-O2"
            export CGO_LDFLAGS="-O2"
          '';

          packages = with pkgs; [
            go_latest
            gotools
            golangci-lint
            gopls

            tailwindcss_4
            tailwindcss-language-server
            air
            templ

            goreleaser
          ];
        };
      }
    );
  };
}
