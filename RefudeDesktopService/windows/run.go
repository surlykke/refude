package windows

import (
	"log"
)

func Run() {
	in.Listen(0)
	updateWindows()

	for {
		if event, err := in.NextEvent(); err != nil {
			log.Println("Error retrieving next event:", err)
		} else {
			handle(event)
		}
	}
}
