let
  pkgs = import <nixpkgs> {};
in
pkgs.mkShell {
	buildInputs = with pkgs; [
		go
		gopls
		gotools
		go-tools
		pkg-config
		gtk4
		gtk4-layer-shell
		pkg-config
	];
}

