self: super: {
  p4-fusion = self.callPackage ./p4-fusion.nix { };

  universal-ctags = self.pkgsStatic.callPackage ./ctags.nix {
    universal-ctags = super.pkgsStatic.universal-ctags;
    # static python is a hassle, and its only used for docs here so we dont care about
    # it being static or not
    python3 = super.python3;
  };

  comby = self.callPackage ./comby.nix {
    # comby = if self.stdenv.hostPlatform.isMacOS then super.comby else super.pkgsMusl.comby;
  };

  nodejs = self.callPackage ./nodejs.nix { };
}
