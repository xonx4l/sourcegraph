let
  rust_overlay = (import (builtins.fetchTarball "https://github.com/oxalica/rust-overlay/archive/master.tar.gz"));

  pkgs = (import <nixpkgs> {
    overlays = [ rust_overlay ];
    config = {
      allowUnfree = true;
    };
    system = "aarch64-darwin";
  });
  # cross = (import <nixpkgs> {
  #   crossSystem = {
  #     config = "aarch64-apple-darwin";
  #     platform = {};
  #     xcodePlatform = "MacOSX";
  #     };
  #   });
  apple_libs = with pkgs.pkgsx86_64Darwin.darwin; [
    apple_sdk.frameworks.Carbon
    apple_sdk.frameworks.WebKit
  ];
  rust-toolchain = pkgs.pkgsx86_64Darwin.rust-bin.fromRustupToolchain {
      channel="stable";
      profile="default";
      components = [
        "rust-std"
        "cargo"
        "rustc"
      ];
      };

in with pkgs.pkgsx86_64Darwin; mkShell {
  nativeBuildInputs = [
    pkg-config
    rust-toolchain

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

  buildInputs =  [apple_libs];

}

