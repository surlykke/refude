#!/usr/bin/env bash
tmp=`mktemp`

cat<< EOF > $tmp
/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

EOF

cat $1 >> $tmp
mv $tmp $1
