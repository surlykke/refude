package icons

import (
	"github.com/surlykke/RefudeServices/lib/xdg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CopyIcons(iconName string, iconThemePath string) {
	if iconName == "" || iconThemePath == "" {
		return
	}

	var pngName = iconName + ".png"
	var xpmName = iconName + ".xpm"
	var svgName = iconName + ".svg"

	var sessionIconDir = xdg.RuntimeDir + "/org.refude.icon-service-session-icons/"
	if !strings.HasSuffix(iconThemePath, "/") {
		iconThemePath = iconThemePath + "/"
	}

	var filesCopied = 0
	filepath.Walk(iconThemePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Error descending into", path)
			return err
		}
		if !info.IsDir() && (info.Name() == pngName || info.Name() == xpmName || info.Name() == svgName) {
			relPath := path[len(iconThemePath):]
			if len(relPath) > len(info.Name()) {
				destDir := sessionIconDir + relPath[0:len(relPath)-len(info.Name())-1]
				if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
					return err
				}
			}

			destPath := sessionIconDir + relPath

			r, err := os.Open(path)
			if err != nil {
				log.Println("Error reading file:", err)
				return err
			}
			defer r.Close()

			w, err := os.Create(destPath)
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
		}
		return nil
	})

	if filesCopied > 0 {
		if _, err := os.Create(sessionIconDir + "/marker"); err != nil {
			log.Println("Error updating marker:", err)
		}
	}
}
