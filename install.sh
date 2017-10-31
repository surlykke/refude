#!/usr/bin/env bash
thisdir=$(dirname $(realpath $0))
cd $thisdir
rm -rf ${thisdir}/dist/*

[[ -n "$PREFIX" ]] || PREFIX=$HOME/.local
REFUDEDIR=${PREFIX}/share/refude
mkdir -p ${REFUDEDIR}

reactapps="panel/refudePanel do/refudeDo appchooser/refudeAppChooser"
for app in $reactapps; do
	appdir=$thisdir/`dirname $app`
	echo "building $appdir"
	cd $appdir
	gulp || exit 1
	cp -R $thisdir/dist/`dirname $app` ${PREFIX}/share/refude
	ln -sf ${PREFIX}/share/refude/$app ${PREFIX}/bin
    for desktopfile in ${PREFIX}/share/refude/`dirname $app`/*.desktop; do
        ln -sf $desktopfile ${PREFIX}/share/applications
    done
done

