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
  rust-toolchain = pkgs.rust-bin.fromRustupToolchain {
      channel="stable";
      profile="default";
      components = [
        "rust-std"
        "cargo"
        "rustc"
      ];
      targets=[
        "aarch64-apple-darwin"
        "x86_64-apple-darwin"
      ];
      };
  apple_libs = with pkgs.darwin; [
    libiconv
    libobjc
    apple_sdk.frameworks.CoreGraphics
    apple_sdk.frameworks.Foundation
    apple_sdk.frameworks.CoreFoundation
    apple_sdk.frameworks.CoreVideo
    apple_sdk.frameworks.AppKit
    apple_sdk.frameworks.QuartzCore
    apple_sdk.frameworks.Security
    apple_sdk.frameworks.WebKit
  ];

in with pkgs; mkShell {
  nativeBuildInputs = [
    pkg-config
    rust-toolchain
    nodejs-16_x
    (nodejs-16_x.pkgs.pnpm.override {
      version = "8.1.0";
      src = fetchurl {
        url = "https://registry.npmjs.org/pnpm/-/pnpm-8.1.0.tgz";
        sha512 = "sha512-e2H73wTRxmc5fWF/6QJqbuwU6O3NRVZC1G1WFXG8EqfN/+ZBu8XVHJZwPH6Xh0DxbEoZgw8/wy2utgCDwPu4Sg==";
      };
    })
    nodePackages.typescript
    go_1_20
  ];

  buildInputs =  [apple_libs];

}

