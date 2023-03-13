package windows

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/windows/x11"
)

func GetResourceRepo() resource.ResourceRepo {
	return x11.Windows
}

func Run() {
	x11.Run()
}

func GetPaths() []string {
	return x11.Windows.GetPaths()
}

func RaiseAndFocusNamedWindow(name string) bool {
	return x11.RaiseAndFocusNamedWindow(name)
}


