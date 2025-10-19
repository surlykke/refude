let
  # You could also pin Nixpkgs to a specific version here; that is not the
  # subject of today's lesson.
  pkgs = import <nixpkgs> { };
  # This empty attribute set is for parameters to Nixpkgs one might wish
  # to configure.
in
pkgs.callPackage ./package.nix { 
}
# And this empty attribute set is for providing dependencies or configuration
# parameters to your package file that aren't automatically selected from
# Nixpkgs. This callPackage call is kinda the same thing as writing:
#     import ./package.nix {
#       inherit (pkgs) lib buildGoModule fetchFromGitHub;
#     }
# but it automatically injects the named dependencies.
