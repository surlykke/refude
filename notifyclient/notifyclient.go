package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"net/http"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

var a fyne.App

func main() {
	a = app.New()
	var window = a.NewWindow("Refude notifier")
	var	imgContainer = container.NewVBox()
	var textContainer = container.NewVBox()
	window.SetContent(container.NewHBox(imgContainer, textContainer))
	go watchNotification(imgContainer, textContainer, os.Args[1])
	window.ShowAndRun()
}


func watchNotification(imgContainer, textContainer *fyne.Container, notificationId string) {
	fmt.Println("watch..")
	getDataOrExit(imgContainer, textContainer, notificationId)
	for range time.Tick(time.Second) {
		getDataOrExit(imgContainer, textContainer, notificationId)
	}

}

type notification struct {
	Title   string `json:"title"`
	Coment  string `json:"comment"`
	Dummy   string `json:"dummy"`
	Icon    string `json:"icon"`
	Expires time.Time
	Created time.Time
	Deleted bool
	Urgency string 
}

func getDataOrExit(imgContainer, textContainer *fyne.Container, nId string) {
	fmt.Println("getData")
	var n notification
	var client = &http.Client{Timeout: 1 * time.Second}
	if response, err := client.Get("http://localhost:7938/notification/" + nId); err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		if response.StatusCode > 299 {
			os.Exit(0)
		} else if err := json.NewDecoder(response.Body).Decode(&n); err != nil {
			fmt.Println(err)
			os.Exit(1)
		} else if stale(n) {
			os.Exit(0)
		} else {
			fmt.Println("Have notification:", n)
			imgContainer.RemoveAll()
			if n.Icon != "" {
				if iconResource, err := fyne.LoadResourceFromURLString(n.Icon); err == nil {
					var icon = canvas.NewImageFromResource(iconResource)
					icon.FillMode = canvas.ImageFillContain
					icon.SetMinSize(fyne.NewSize(32, 32))
					fmt.Println("Add icon")
					imgContainer.Add(icon)
				}
			}
			textContainer.RemoveAll()
			var subject = canvas.NewText(n.Title, color.White)
			subject.TextSize = 20
			textContainer.Add(subject)
			var body = canvas.NewText(n.Coment, color.White)
			textContainer.Add(body)
			fmt.Println("Done")
		}
	}
}

func stale(n notification) bool {
	fmt.Println("stale, urgency:", n.Urgency, ", created:", n.Created)
	return n.Deleted ||
		time.Now().After(n.Expires) ||
		n.Urgency == "low" && time.Now().After(n.Created.Add(4*time.Second)) ||
		n.Urgency == "normal" && time.Now().After(n.Created.Add(10*time.Second))
}

