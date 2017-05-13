#!/bin/bash
cd `dirname $0`

for dir in RefudeDesktopService RefudeIconService RefudePowerService RefudeWmService RefudeXdgOpen; do
	(cd $dir && echo "Building $dir" && go install) || exit 1
done

cp runRefude.sh ../../../../bin 
