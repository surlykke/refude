#!/usr/bin/env bash
thisdir=$(dirname $(realpath $0))

cd ~/bin
angularapps="panel/refudePanel power/refudePower appconfig/refudeAppConfig connman/refudeConnman"

for app in $angularapp; do
	ln -sf $thisdir/$app
done

reactapps="do/refudeDo"
for app in $reactapps; do
	appdir=`dirname $app`
	echo "building $appdir"
	(cd $thisdir/$appdir && gulp) && ln -sf $thisdir/dist/$app
done

