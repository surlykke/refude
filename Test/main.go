package main

import (
	"time"
	"fmt"
)

func main() {
	var ticker = time.Tick(time.Second);

	for now := range ticker {
		fmt.Println("tick!", now)
	}
}
