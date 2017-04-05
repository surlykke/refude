#!/usr/bin/env bash
thisdir=$(dirname $(realpath $0))
applist="do/refudeDo panel/refudePanel power/refudePower appconfig/refudeAppConfig connman/refudeConnman"

cd ~/bin
for app in $applist; do
   ln -sf $thisdir/$app
done

