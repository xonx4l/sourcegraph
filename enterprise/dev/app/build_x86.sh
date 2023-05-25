#! /usr/bin/env nix-shell
#! nix-shell app-deps.nix -i bash

which rustc
which cargo
rustc -vV

pnpm tauri build
pnpm tauri build --target x86_64-apple-darwin

