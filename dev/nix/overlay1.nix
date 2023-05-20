{ lib }: lib.composeManyExtensions [
  (import ./ctags.nix)
  (import ./p4-fusion.nix)
  (import ./comby.nix)
  (import ./nodejs.nix)
]
