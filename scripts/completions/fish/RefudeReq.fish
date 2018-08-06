function services
    echo services
	ls $XDG_RUNTIME_DIR/org.refude.*service | sed -E 's/.*org\.refude\.//'
end

function paths
	RefudeReq (commandline -op)[2] GET /links 2>/dev/null | grep -Po '"/[^"]*"\s*' | sed -e 's/"\(.*\)"\s*/\1/g' 2>/dev/null
end

complete -c RefudeReq -n "test (count (commandline -opc)) -eq 1" -a "(services)" -f
complete -c RefudeReq -n "test (count (commandline -opc)) -eq 2" -a "GET POST PATCH DELETE" -f
complete -c RefudeReq -n "test (count (commandline -opc)) -eq 3" -a "(paths)" -f
