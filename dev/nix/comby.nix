{ pkgs
, pkgsStatic
, pkgsMusl
, lib
, hostPlatform
  # , comby
, zlib
, libev
, gmp
, ocamlPackages
, sqlite
}:
let
  inherit (import ./util.nix { inherit lib; }) makeStatic unNixifyDylibs;
  combyBuilder = ocamlPkgs: systemDepsPkgs:
    (ocamlPkgs.comby.override {
      sqlite = systemDepsPkgs.sqlite;
      zlib = systemDepsPkgs.zlib.static or systemDepsPkgs.zlib;
      libev = (makeStatic (systemDepsPkgs.libev)).override { static = false; };
      gmp = makeStatic systemDepsPkgs.gmp;
      ocamlPackages = ocamlPkgs.ocamlPackages.overrideScope' (self: super: {
        ocaml_pcre = super.ocaml_pcre.override {
          pcre = makeStatic systemDepsPkgs.pcre;
        };
        ssl = super.ssl.override {
          openssl = (makeStatic systemDepsPkgs.openssl).override { static = true; };
        };
      });
    });
in
if hostPlatform.isMacOS then
  unNixifyDylibs pkgs (combyBuilder pkgs pkgsStatic)
else
  (combyBuilder pkgsMusl pkgsStatic).overrideAttrs (_: {
    postPatch = ''
      cat >> src/dune <<EOF
      (env (release (flags  :standard -ccopt -static)))
      EOF
    '';
  })
