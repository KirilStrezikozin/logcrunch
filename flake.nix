{
  description = "Logcrunch - Profiling via Structured Logs";

  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixos-25.05/nixexprs.tar.xz";
    flake-utils.url = "github:numtide/flake-utils";

    # Used for shell.nix
    flake-compat = {
      url = "github:edolstra/flake-compat";
      flake = false;
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        # Development shell
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            golangci-lint
            tailwindcss_4
            templ
            just
            air
            pre-commit

            nixd
            gopls
            golangci-lint-langserver
            htmx-lsp
            tree

            nixfmt-rfc-style

            nodejs_latest
            watchman
          ];
        };

        # For compatibility with older versions of the `nix` binary
        devShell = self.devShells.${system}.default;

        # Formatter to use with the `nix fmt` command.
        formatter = pkgs.nixfmt-tree;
      }
    );
}
