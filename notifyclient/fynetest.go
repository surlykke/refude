package main

/*import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
)

const url = "http://localhost:7938/icon?name=%2Fsnap%2Fcode%2F159%2Fmeta%2Fgui%2Fvscode.png"

var a fyne.App 

func dummy() {
	fmt.Println("Start")
	//"http://localhost:7938/icon?name=firefox"
	time.AfterFunc(3*time.Second, showWindow)
	time.AfterFunc(5*time.Second, func() { w.Hide() })
	time.AfterFunc(7*time.Second, func() { w.Show() })
	a = app.NewWithID("54546565")
	a.Run()

}

func closeWindow() {
	w.Hide()
}

func showWindow() {
	var oldWindow = w
		
	w = a.NewWindow("Refude notifier")
	imgUrl, err := storage.ParseURI(url)
	if err != nil {
		panic(err)
	}
	img := canvas.NewImageFromURI(imgUrl)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(32, 32))
	imgText := canvas.NewText(" ", color.Opaque)
	imgContainer := container.NewVBox(imgText)
	text1 := canvas.NewText("Subject", color.White)
	text1.TextSize = 24
	text2 := canvas.NewText("lksdjf lkjsfd lksdjfldsfkjdsf  lkdsjf lkjdf lkjds lskfdsjfdjweurn lsdkfj", color.White)
	textContainer := container.NewVBox(text1, text2)
	mainContainer := container.NewHBox(imgContainer, textContainer)
	w.SetContent(mainContainer)
	//time.AfterFunc(40*time.Second, func() {w.Close()})
	w.Show()
	if oldWindow != nil {
		oldWindow.Close()
	}
}*/
