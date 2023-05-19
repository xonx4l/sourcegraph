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
        callPackageWith' = (import ./dev/nix/util.nix { inherit (pkgs) lib; }).callPackageWith;
        overlays = callPackageWith' pkgs ./dev/nix/overlay.nix { };
        # pkgs' = pkgs.lib.fold (a: b: b.extend a) pkgs (builtins.attrValues overlays);
        pkgs' = import nixpkgs { inherit system; overlays = builtins.attrValues overlays; };
          # pkgs' = pkgs.extend (self: super: overlays { inherit self super; });
          in
          {
          legacyPackages = pkgs';
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
