{ lib }:
{
  # utility function to add some best-effort flags for emitting static objects instead of dynamic
  makeStatic = pkg:
    let
      auto = builtins.intersectAttrs pkg.override.__functionArgs { withStatic = true; static = true; enableStatic = true; enableShared = false; };
      overridden = pkg.overrideAttrs (oldAttrs: {
        dontDisableStatic = true;
      } // lib.optionalAttrs (!(oldAttrs.dontAddStaticConfigureFlags or false)) {
        configureFlags = (oldAttrs.configureFlags or [ ]) ++ [ "--disable-shared" "--enable-static" "--enable-shared=false" ];
      });
    in
    overridden.override auto;

  # doesn't actually change anything in practice, just makes otool -L not display nix store paths for libiconv and libxml.
  # they exist in macos dydl cache anyways, so where they point to is irrelevant. worst case, this will let you catch earlier
  # when a library that should be statically linked or that isnt in dydl cache is dynamically linked.
  unNixifyDylibs = pkgs: drv:
    drv.overrideAttrs (oldAttrs: {
      postFixup = with pkgs; (oldAttrs.postFixup or "") + lib.optionalString pkgs.hostPlatform.isMacOS ''
        for bin in $(${findutils}/bin/find $out/bin -type f); do
          for lib in $(otool -L $bin | ${coreutils}/bin/tail -n +2 | ${coreutils}/bin/cut -d' ' -f1 | ${gnugrep}/bin/grep nix); do
            install_name_tool -change "$lib" "@rpath/$(basename $lib)" $bin
          done
        done
      '';
    });

  # same as callPackageWith but doesn't apply makeOverridable[1]. See [2] for nixpkgs exemplar.
  # [1] https://sourcegraph.com/github.com/NixOS/nixpkgs@1a6a0923e57d9f41bcc3e2532a7f99943a3fbaeb/-/blob/lib/customisation.nix?L78
  # [2] https://sourcegraph.com/github.com/NixOS/nixpkgs@1a6a0923e57d9f41bcc3e2532a7f99943a3fbaeb/-/blob/pkgs/development/beam-modules/lib.nix?L8
  callPackageWith = autoArgs: fn: args:
    let
      f = if lib.isFunction fn then fn else import fn;
      auto = builtins.intersectAttrs (lib.functionArgs f) autoArgs;
    in
    f (auto // args);
}
