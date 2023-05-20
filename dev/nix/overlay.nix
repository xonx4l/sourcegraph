{ pkgs, lib, hostPlatform }:
let
  expandOverlay = name: value: (self: super: {
    name = value { inherit self super; };
  });
in
builtins.mapAttrs expandOverlay
  {
    # universal-ctags = { self, super }: super.lib.callPackageWith pkgs ./ctags.nix { };

    # comby = { self, super }: super.lib.callPackageWith pkgs ./comby.nix { };

    # nodejs = { self, super }: super.lib.callPackageWith pkgs ./nodejs.nix { };
  } // lib.optionalAttrs (hostPlatform.system != "aarch64-linux") {
  p4-fusion = { self, super }: super.callPackage ./p4-fusion.nix { };
}
