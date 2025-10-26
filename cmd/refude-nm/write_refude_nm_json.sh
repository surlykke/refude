#!/usr/bin/env bash
if [[ "-h" == "$1" ]]; then
cat <<-EOF

usage:

    write_refude_nm_json.sh > /path/chome/configdir/NativeMessagingHosts/org.refude.native_messaging.json

More often than not, you'll find chome config dir at $HOME/.config/google-chrome

EOF
elif [[ "" != "$1" ]]; then
	echo "try: write_refude_nm_json.sh -h"
else
cat <<-EOF
{
  "name": "org.refude.native_messaging",
  "description": "Refude browserextension messaging",
  "path": "$(which refude-nm)",
  "type": "stdio",
  "allowed_origins": ["chrome-extension://lcnmbmoiobgochkfoenopkgnoojgbeio/"]
}
EOF
fi
