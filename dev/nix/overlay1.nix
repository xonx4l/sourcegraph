self: super: {
  p4-fusion = super.callPackage ./p4-fusion.nix { };

  universal-ctags' = super.universal-ctags;
  universal-ctags = super.callPackage ./ctags.nix { };

  comby' = super.comby;
  comby = super.callPackage ./comby.nix { };

  nodejs = super.callPackage ./nodejs.nix { };
}
