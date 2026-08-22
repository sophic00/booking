{
  description = "Development environment for ticket-booking with Go, Next.js (Node.js), and Make";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            # Build tools
            gnumake

            # Go toolchain & tools
            go
            gopls
            golangci-lint
            sqlc

            # Node.js runtime & package manager for Next.js
            nodejs_24
            corepack
          ];

          shellHook = ''
            echo "🚀 Development environment loaded!"
            echo "  • Go:   $(go version 2>/dev/null || echo 'N/A')"
            echo "  • Node: $(node --version 2>/dev/null || echo 'N/A')"
            echo "  • Make: $(make --version 2>/dev/null | head -n 1)"
            echo "  • sqlc: $(sqlc version 2>/dev/null || echo 'N/A')"
          '';
        };
      });
}
