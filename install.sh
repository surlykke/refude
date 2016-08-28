#!/usr/bin/env bash
sharedir="/usr/local/share/refude"
filelist="do bin main.js package.json"
cd $(realpath $(dirname $0))
mkdir -p $sharedir
for file in $filelist; do
	cp -R $file $sharedir
done

chmod 755 $sharedir/bin/*.sh

cd /usr/local/bin
for exe in $sharedir/bin/*.sh; do 
	ln -sf $exe; 
done
