#!/usr/bin/env bash
sharedir="/usr/local/share/refude"
applist="do panel power appconfig"
dirlist="common-js $applist"

cd $(realpath $(dirname $0))

mkdir -p $sharedir
for file in $dirlist; do
	cp -R $file $sharedir
done

chmod 744 $sharedir/*/*.desktop

cd /usr/local/bin
for app in $applist; do 
	chmod a+x $sharedir/$app/runRefude*;
	ln -sf $sharedir/$app/runRefude*; 
done

cd /usr/local/share/applications
for desktop in $sharedir/*/*.desktop; do 
	ln -sf $desktop
done
