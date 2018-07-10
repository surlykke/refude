function services
	ls $XDG_RUNTIME_DIR/org.refude.*service | sed -E 's/.*org\.refude\.//'
end

function paths
	RefudeGET $argv[1] /links | grep -Po '"/[^"]*"\s*' | sed -e 's/"\(.*\)"\s*/\1/g'
end

complete -c RefudeGET -n "test (count (commandline -opc)) -lt 2" -a "(services)" -f 
complete -c RefudeGET -n "test (count (commandline -opc)) -ge 2" -a "(paths (commandline -op)[2] (commandline -op)[3])" -f
