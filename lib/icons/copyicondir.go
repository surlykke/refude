package icons

import (
	"os"
	"log"
	"io"
	"strings"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"path/filepath"
)

func CopyIconDir(dir string) {
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	var destDir = xdg.RuntimeDir + "/org.refude.icon-service-session-icons/"
	var filesCopied = 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Error descending into", path)
			return err
		}
		var relPath = path[len(dir):]
		if info.IsDir() {
			if err := os.MkdirAll(destDir+relPath, os.ModePerm); err != nil {
				return err
			}
		} else if _, err := os.Stat(destDir + relPath); err != nil {
			if os.IsNotExist(err) {
				r, err := os.Open(path);
				if err != nil {
					log.Println("Error reading file:", err)
					return err
				}
				defer r.Close()

				w, err := os.Create(destDir + relPath);
				if err != nil {
					log.Println("Error creating file", err)
					return err
				}
				defer w.Close()

				if _, err := io.Copy(w, r); err != nil {
					log.Println("Error copying file", err)
					return err
				}
				filesCopied++
			} else {
				log.Println("Error stat'ing file", err)
				return err
			}
		}
		return nil
	})
	if filesCopied > 0 {
		if _,err := os.Create(destDir + "/marker"); err != nil {
			log.Println("Error updating marker:", err)
		}
	}
}
