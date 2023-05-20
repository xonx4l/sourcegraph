final: prev:
let
  inherit (import ./util.nix { inherit (prev) lib; }) makeStatic;
  http-parser-static = ((makeStatic prev.pkgsStatic.http-parser).overrideAttrs (oldAttrs: {
    # http-parser makefile is a bit incomplete, so fill in the gaps here
    # to move the static object and header files to the right location
    # https://github.com/nodejs/http-parser/issues/310
    buildFlags = [ "package" ];
    installTargets = "package";
    postInstall = ''
      install -D libhttp_parser.a $out/lib/libhttp_parser.a
      install -D  http_parser.h $out/include/http_parser.h
      ls -la $out/lib $out/include
    '';
  }));
  libiconv-static = makeStatic prev.pkgsStatic.libiconv;
  openssl-static = (prev.pkgsStatic.openssl_1_1.override { static = true; }).dev;
  pcre-static = (makeStatic prev.pkgsStatic.pcre).dev;
in
{
  p4-fusion = prev.pkgsStatic.gccStdenv.mkDerivation rec {
    name = "p4-fusion";
    version = "v1.12";

    srcs = [
      (prev.fetchFromGitHub {
        inherit name;
        owner = "salesforce";
        repo = "p4-fusion";
        rev = "3ee482466464c18e6a635ff4f09cd75a2e1bfe0f";
        hash = "sha256-rUXuBoXuOUanWxutd7dNgjn2vLFvHQ0IgCIn9vG5dgs=";
      })
      (
        if prev.targetPlatform.isMacOS then
          if prev.targetPlatform.isAarch64 then
            prev.fetchzip
              {
                name = "helix-core-api";
                url = "https://cdist2.perforce.com/perforce/r22.2/bin.macosx12arm64/p4api-openssl1.1.1.tgz";
                hash = "sha256-YO7p24PuedTn2pVq/roF2u5zqS6byaG9N2gCbGVrpv0=";
              }
          else
            prev.fetchzip {
              name = "helix-core-api";
              url = "https://cdist2.perforce.com/perforce/r22.2/bin.macosx12x86_64/p4api-openssl1.1.1.tgz";
              hash = "sha256-gaYvQOX8nvMIMHENHB0+uklyLcmeXT5gjGGcVC9TTtE=";
            }
        else if prev.targetPlatform.isLinux then
          prev.fetchzip
            {
              name = "helix-core-api";
              url = "https://cdist2.perforce.com/perforce/r22.2/bin.linux26x86_64/p4api-glibc2.3-openssl1.1.1.tgz";
              hash = "sha256-JkWG4ImrTzN0UuSMelG8zsH7YRlL1mXs9lpB5GptUb4=";
            }
        else throw "unsupported platform ${prev.stdenv.targetPlatform.parsed.kernel.name}"
      )
    ];

    sourceRoot = name;

    nativeBuildInputs = with prev; [
      patchelf
      pkg-config
      cmake
    ];

    buildInputs = with prev.pkgsStatic; [
      zlib
      zlib.dev
      http-parser-static
      pcre-static
      openssl-static
    ] ++ lib.optional targetPlatform.isMacOS [
      # iconv is bundled with glibc and apparently only needed for osx
      # https://sourcegraph.com/github.com/salesforce/p4-fusion@3ee482466464c18e6a635ff4f09cd75a2e1bfe0f/-/blob/vendor/libgit2/README.md?L178:3
      libiconv-static
      darwin.apple_sdk.frameworks.CFNetwork
      darwin.apple_sdk.frameworks.Cocoa
    ];

    # copy helix-core-api stuff into the expected directories, and statically link libstdc++
    preBuild = let dir = if prev.targetPlatform.isMacOS then "mac" else "linux"; in
      ''
        mkdir -p $NIX_BUILD_TOP/$sourceRoot/vendor/helix-core-api/${dir}
        cp -R $NIX_BUILD_TOP/helix-core-api/* $NIX_BUILD_TOP/$sourceRoot/vendor/helix-core-api/${dir}

        sed -i "s/target_link_libraries(p4-fusion PUBLIC/target_link_libraries(p4-fusion PUBLIC -static-libstdc++/" \
          $NIX_BUILD_TOP/$sourceRoot/p4-fusion/CMakeLists.txt
      '';

    cmakeFlags = [
      # we want to statically link
      "-DBUILD_SHARED_LIBS=OFF"
      # Copied from upstream, where relevant
      # https://sourcegraph.com/github.com/salesforce/p4-fusion@3ee482466464c18e6a635ff4f09cd75a2e1bfe0f/-/blob/generate_cache.sh?L7-21
      "-DUSE_SSH=OFF"
      "-DUSE_HTTPS=OFF"
      "-DBUILD_CLAR=OFF"
      # salesforce don't link against GSSAPI in CI, so I won't either
      "-DUSE_GSSAPI=OFF"
      # prefer nix-provided http-parser instead of bundled
      "-DUSE_HTTP_PARSER=system"
    ];

    postInstall = ''
      mkdir -p "$out/bin"
      cp p4-fusion/p4-fusion "$out/bin/p4-fusion"
    '';

    meta = {
      homepage = "https://github.com/salesforce/p4-fusion";
      platforms = [ "x86_64-darwin" "aarch64-darwin" "x86_64-linux" ];
      license = prev.lib.licenses.bsd3;
    };
  };
}
