package lib

import (
	"fmt"
	"os/exec"
	"log"
)

func RunCmd(argv []string) {
	fmt.Println("runCmd")
	for i := 0; i < len(argv); i++ {
		fmt.Println(i, ":", argv[i])
	}
	var cmd = exec.Command(argv[0], argv[1:]...)

	cmd.Dir = Home
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		log.Println(err.Error())
		return
	}

	go cmd.Wait() // TODO Rather use double fork
}

