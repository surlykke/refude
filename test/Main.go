// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func slashes(path string) int64 {
	var num = int64(0)
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			num++
		}
	}
	return num
}

func main() {
	var start = time.Now()
	var count = 0
	filepath.Walk("/home/surlykke", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			count++
		}

		return nil
	})
	var end = time.Now()
	fmt.Println("Found", count, "files in", end.Sub(start))
}
