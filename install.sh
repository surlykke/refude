#!/usr/bin/env bash
thisdir=$(dirname $(realpath $0))

[[ -n "$PREFIX" ]] || PREFIX=$HOME/.local
REFUDEDIR=${PREFIX}/share/refude
BINDIR=${PREFIX}/bin
mkdir -p ${REFUDEDIR}

reactapps="panel/refudePanel do/refudeDo appchooser/refudeAppChooser"
for app in $reactapps; do
	appdir=$thisdir/`dirname $app`
	echo "building $appdir"
	cd $appdir
	gulp || exit 1
	cp -R $thisdir/dist/`dirname $app` ${PREFIX}/share/refude
	ln -sf ${PREFIX}/share/refude/$app $BINDIR
done

