{ pkgs
, lib
, fetchFromGitHub
, python3
, universal-ctags
, pcre2
, libyaml
, jansson
}:
let
  inherit (import ./util.nix { inherit lib; }) makeStatic unNixifyDylibs;
  pcre2-static = makeStatic pcre2;
  libyaml-static = makeStatic libyaml;
  jansson-static = jansson.overrideAttrs (oldAttrs: {
    cmakeFlags = [ "-DJANSSON_BUILD_SHARED_LIBS=OFF" ];
  });
in
unNixifyDylibs pkgs ((universal-ctags.override {
  inherit python3;
  pcre2 = pcre2-static;
  libyaml = libyaml-static;
  jansson = jansson-static;
}).overrideAttrs (oldAttrs: {
  version = "5.9.20220403.0";
  src = fetchFromGitHub {
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
  postFixup = (oldAttrs.postFixup or "") + ''
    # unideal -f here, theres a weird eval loop going on requiring
    # 1) the package function containing the params passed to .override above and
    # 2) operations here being idempotent
    ln -sf $out/bin/ctags $out/bin/universal-ctags
  '';
}))
