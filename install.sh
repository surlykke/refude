#!/usr/bin/env bash
sharedir="/usr/local/share/refude"
filelist="bin common-js appconfig do panel power createwin.js main.js package.json"
cd $(realpath $(dirname $0))
mkdir -p $sharedir
for file in $filelist; do
	cp -R $file $sharedir
done

chmod 755 $sharedir/bin/*.sh 
chmod 744 $sharedir/*/*.desktop

cd /usr/local/bin
for exe in $sharedir/bin/*.sh; do 
	ln -sf $exe; 
done

cd /usr/local/share/applications
for desktop in $sharedir/*/*.desktop; do 
	ln -sf $desktop
done
