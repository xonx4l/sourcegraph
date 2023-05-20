{
  description = "The Sourcegraph developer environment & packages Nix Flake";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        pkgs' = pkgs.extend (import ./dev/nix/overlay1.nix);
      in
      {
        legacyPackages = builtins.removeAttrs pkgs' [ "ctags" ];
        devShells.default = pkgs'.callPackage ./shell.nix { };
        packages = {
          inherit (pkgs') universal-ctags comby nodejs;
        } // pkgs.lib.optionalAttrs (pkgs.hostPlatform.system != "aarch64-linux") {
          inherit (pkgs') p4-fusion;
        };
        #     bazel-fhs = (pkgs.buildFHSEnv {
        #       name = "bazel";
        #       runScript = "bash";
        #       targetPkgs = pkgs: (with pkgs; [
        #         bazel_6
        #         zlib.dev
        #       ]);
        #     }).env;
        #   }
        # );

        formatter = pkgs.nixpkgs-fmt;
      });
}
