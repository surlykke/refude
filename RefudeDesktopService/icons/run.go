package icons

import "github.com/surlykke/RefudeServices/lib/image"

var Icons = MakeIconCollection()

var BasedirSink = make(chan string)
var ARGBIconSink = make(chan image.ARGBIcon)

func Run() {
	Icons.collect()
	for {
		select {
		case basedir := <-BasedirSink:
			Icons.addBasedir(basedir)
		case argbIcon := <-ARGBIconSink:
			Icons.addARGBIcon(argbIcon)

		}
	}
}
