let
  # rust_overlay = (import (builtins.fetchTarball "https://github.com/oxalica/rust-overlay/archive/master.tar.gz"));

  # rust_custom = (self: super:
  #   {
  #     rust-stable = super.rust-bin.stable.latest.default.overrideAttrs (old: {
  #       propagatedBuildInputs = [ ];
  #       depsHostHostPropagated = [ ];
  #       depsTargetTargetPropagated = [ ];
  #     });
  #   });

  pkgs = (import (builtins.fetchTarball "https://github.com/NixOS/nixpkgs/archive/db38340b4d9a987db4cd4e46a537851d27bc6f44.tar.gz") {
    # overlays = [ rust_overlay rust_custom ];
    config = {
      allowUnfree = true;
    };
    system = "aarch64-darwin";
  });

  apple_libs = with pkgs.pkgsx86_64Darwin.darwin.apple_sdk_11_0; [
    frameworks.AppKit
    frameworks.Foundation
    frameworks.CoreFoundation
    frameworks.Carbon
    frameworks.WebKit
  ];

  mkClangShell = pkgs.pkgsx86_64Darwin.mkShell.override { stdenv = pkgs.pkgsx86_64Darwin.darwin.apple_sdk_11_0.clang13Stdenv; };

in
with pkgs.pkgsx86_64Darwin; mkClangShell {
  nativeBuildInputs = [
    pkg-config
    rustc
    cargo
    git
    openssh
    which

    nodejs-18_x
    (nodejs-18_x.pkgs.pnpm.override {
      version = "8.1.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/pnpm/-/pnpm-8.1.0.tgz";
        sha512 = "sha512-e2H73wTRxmc5fWF/6QJqbuwU6O3NRVZC1G1WFXG8EqfN/+ZBu8XVHJZwPH6Xh0DxbEoZgw8/wy2utgCDwPu4Sg==";
      };
    })
    nodejs-18_x.pkgs.typescript

    go_1_20
  ] ++ apple_libs;

  buildInputs = [
    libiconv
  ];
}

