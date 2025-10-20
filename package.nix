# First, the named parameters.
{
  # you almost always want to depend on lib
  lib,

  # these are Nixpkgs functions that your package uses
  buildGoModule,
  fetchFromGitHub,
  pkg-config,
  kdePackages,
  gtk4-layer-shell,
  gtk4,
  # more dependencies would go here...
}: # this means end of named parameters

# Now, the definition of your package.
# This should be something that produces a derivation, not
# a string or a raw attribute set or anything else.
# buildGoModule is a function that returns a derivation, so
# you want `buildGoModule ...` here, not `{ pet = ...; }` here;
# the latter is an attribute set.

buildGoModule rec {
  pname = "refude";
  version = "0.1";

  src = ./.;
#  src = fetchFromGitHub {
#    owner = "surlykke";
#    repo = "";
#    rev = "v${version}";
#    hash = "sha256-Gjw1dRrgM8D3G7v6WIM2+50r4HmTXvx0Xxme2fH9TlQ=";
#  };

	#  buildInputs = [ pkg-config ] ++ libs;

  # this hash is updated from the example, which seems to be out of date
  vendorHash = "sha256-tfcuyrrxxw5RF33vo2qXH7jnSiYXbKpw3UrNtrsRc30=";

  subPackages = [ "cmd/refude-server" "cmd/refuc" "cmd/refude-nm" ];

  nativeBuildInputs = [ pkg-config ];

  buildInputs = [ kdePackages.wayland gtk4 gtk4-layer-shell ];

  postInstall = 
  ''
	mkdir -p $out/bin
	cp cmd/refude-server/runRefude.sh $out/bin
  '';

  meta = {
    description = "Window switcher and more";
    homepage = "https://github.com/surlykke/refude";
    license = lib.licenses.gpl2;
    maintainers = with lib.maintainers; [ surlykke ];
  };
}
