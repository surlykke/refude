{
  lib,
  buildGoModule,
  fetchFromGitHub,
  pkg-config,
  kdePackages,
  gtk4-layer-shell,
  gtk4,
}: 
buildGoModule rec {
  pname = "refude";
  version = "0.1";

  src = lib.cleanSource ./.;
  vendorHash = "sha256-tfcuyrrxxw5RF33vo2qXH7jnSiYXbKpw3UrNtrsRc30=";
  subPackages = [ "cmd/refude-server" "cmd/refuc" "cmd/refude-nm" ];
  nativeBuildInputs = [ pkg-config ];
  buildInputs = [ kdePackages.wayland gtk4 gtk4-layer-shell ];
  postInstall = 
  ''
	mkdir -p $out/bin
	cp cmd/refude-server/runRefude.sh $out/bin
	mkdir -p $out/.local/share/bash/completions
	cp cmd/refuc/completions/bash/* $out/.local/share/bash/completions
	mkdir -p $out/.local/share/fish/completions
	cp cmd/refuc/completions/fish/* $out/.local/share/fish/completions
	mkdir -p $out/share/icons/hicolor
	cp -R internal/refudeicons/* $out/share/icons/hicolor/
  '';

  meta = {
    description = "Window switcher and more";
    homepage = "https://github.com/surlykke/refude";
    license = lib.licenses.gpl2;
    maintainers = with lib.maintainers; [ surlykke ];
  };
}
