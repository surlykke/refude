function services
	ls $XDG_RUNTIME_DIR/org.refude.* | sed -E 's/.*org\.refude\.//' 
end

function paths
	echo paths $argv1 $argv2 > ~/tmp/log.txt
	set base (echo $argv[2] | sed -e 's/[^/]*$//') 
	set elements (RefudeGET $argv[1] $base | sed -e 's/[][",]/ /g' | string trim | tr -s ' ' \n)
	for element in $elements
		echo {$base}$element
	end
end

complete -c RefudePOST -n "test (count (commandline -opc)) -lt 2" -a "(services)" -f 
complete -c RefudePOST -n "test (count (commandline -opc)) -ge 2" -a "(paths (commandline -op)[2] (commandline -op)[3])" -f
