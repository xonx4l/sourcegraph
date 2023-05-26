let
  rust_overlay = (import (builtins.fetchTarball "https://github.com/oxalica/rust-overlay/archive/master.tar.gz"));

  rust_custom = (self: super:
  {
    rust-stable = super.rust-bin.stable.latest.default.overrideAttrs (old: {
      propagatedBuildInputs = [];
      depsHostHostPropagated = [];
      depsTargetTargetPropagated = [];
    });
  });

  pkgs = (import <nixpkgs> {
    overlays = [ rust_overlay rust_custom ];
    config = {
      allowUnfree = true;
    };
    system = "aarch64-darwin";
  });

  apple_libs = with pkgs.pkgsx86_64Darwin.darwin; [
    apple_sdk.frameworks.AppKit
    apple_sdk.frameworks.Foundation
    apple_sdk.frameworks.CoreFoundation
    apple_sdk.frameworks.Carbon
    apple_sdk.frameworks.WebKit
  ];

  mkClangShell = pkgs.pkgsx86_64Darwin.mkShell.override { stdenv = pkgs.pkgsx86_64Darwin.clang13Stdenv; };

in with pkgs.pkgsx86_64Darwin; mkClangShell {
  nativeBuildInputs = [
    pkg-config
    rust-stable

    pkgs.nodejs-16_x
    (pkgs.nodejs-16_x.pkgs.pnpm.override {
      version = "8.1.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/pnpm/-/pnpm-8.1.0.tgz";
        sha512 = "sha512-e2H73wTRxmc5fWF/6QJqbuwU6O3NRVZC1G1WFXG8EqfN/+ZBu8XVHJZwPH6Xh0DxbEoZgw8/wy2utgCDwPu4Sg==";
      };
    })
    pkgs.nodePackages.typescript

    go_1_20
  ];

  buildInputs =  [
    libiconv
  ];
  #CC="${clang13Stdenv.cc.cc}/bin/clang";
  CARGO_TARGET_X86_64_APPLE_DARWIN_LINKER = "${clang13Stdenv.cc.cc}/bin/clang";
}

