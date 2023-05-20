final: prev:
let
  inherit (import ./util.nix { inherit (prev) lib; }) makeStatic unNixifyDylibs;
  pcre2-static = makeStatic prev.pkgsStatic.pcre2;
  libyaml-static = makeStatic prev.pkgsStatic.libyaml;
  jansson-static = prev.pkgsStatic.jansson.overrideAttrs (oldAttrs: {
    cmakeFlags = [ "-DJANSSON_BUILD_SHARED_LIBS=OFF" ];
  });
in
{
  universal-ctags = unNixifyDylibs prev.pkgs
    ((prev.pkgsStatic.universal-ctags.override {
      inherit (prev) python3;
      pcre2 = pcre2-static;
      libyaml = libyaml-static;
      jansson = jansson-static;
    }).overrideAttrs (oldAttrs: {
      version = "5.9.20220403.0";
      src = prev.fetchFromGitHub {
        owner = "universal-ctags";
        repo = "ctags";
        rev = "f95bb3497f53748c2b6afc7f298cff218103ab90";
        sha256 = "sha256-pd89KERQj6K11Nue3YFNO+NLOJGqcMnHkeqtWvMFk38=";
      };
      # disable checks, else we get `make[1]: *** No rule to make target 'optlib/cmake.c'.  Stop.`
      doCheck = false;
      checkFlags = [ ];
      # don't include libintl/gettext
      dontAddExtraLibs = true;
      postFixup = builtins.trace "uhoh" (oldAttrs.postFixup or "") + ''
        # unideal -f here, theres a weird double-eval going on here
        ln -sf $out/bin/ctags $out/bin/universal-ctags
      '';
    }));
}
