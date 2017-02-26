package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(time.Now(), "Starting")
	d := DesktopService{}
	d.Start()

}
