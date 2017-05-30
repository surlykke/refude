package main

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/service"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/stringlist"
	"github.com/surlykke/RefudeServices/lib/argb"
)

var ItemFields = []string{ "Id", "Category", "Status", "Title", "ItemIsMenu", "IconName", "AttentionIconName", "OverlayIconName", "AttentionMovieName", "IconUrl", "AttentionIconUrl", "OverlayIconUrl", "ToolTip" }

type Item struct {
	Id string
	Category string
	Status string
	Title string
	ItemIsMenu bool

	IconName string
	AttentionIconName string
	OverlayIconName string
	AttentionMovieName string
	IconUrl string
	AttentionIconUrl string
	OverlayIconUrl string
	ToolTip  ToolTip
}

type ToolTip struct {
	Title       string
	Description string
	IconName    string
	IconUrl     string
}

func (item Item) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		stringlist.ServeAsJson(w, r, item)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}


func collectPixMap(dbusValue [][]interface{}) argb.Icon {
	res := make(argb.Icon, 0)
	for _,arr := range(dbusValue) {
		for len (arr) > 2 {
			width := arr[0].(int32)
			height := arr[1].(int32)
			pixels := arr[2].([]byte)
			res = append(res, argb.Img{Width: width, Height: height, Pixels: pixels})
			arr = arr[3:]
		}
	}

	return res
}

func StatusNotifierItem(serviceId string, propUpdates PropChangeChannel) {
	id, err := getId(serviceId)
	if err != nil {
		return
	}
	path := "/items/" + id
	var itemData Item = Item{}

	for update := range propUpdates {
		for key, variant := range update {
			switch key {
			case "Id":
				itemData.Id = variant.Value().(string)
			case "Category":
				itemData.Category = variant.Value().(string)
			case "Status":
				itemData.Status = variant.Value().(string)
			case "Title":
				itemData.Title = variant.Value().(string)
			case "ItemIsMenu":
				itemData.ItemIsMenu = variant.Value().(bool)
			case "IconName":
				itemData.IconName = variant.Value().(string)
			case "AttentionIconName":
				itemData.AttentionIconName = variant.Value().(string)
			case "OverlayIconName":
				itemData.OverlayIconName = variant.Value().(string)
			case "AttentionMovieName":
				itemData.AttentionMovieName = variant.Value().(string)
			case "IconPixmap":
				icon := collectPixMap(variant.Value().([][]interface{}))
				itemData.IconUrl,_ = argb.ServeAsPng(icon)
			case "AttentionIconUrl":
				itemData.AttentionIconUrl = variant.Value().(string)
			case "OverlayIconUrl":
				itemData.OverlayIconUrl = variant.Value().(string)
			case "ToolTip":
				fmt.Println("ToolTip: ", variant.Value())
			}
		}

		fmt.Println("mapping to ", path)
		service.Map(path, itemData)
	}

	service.Unmap(path)
}

// TODO MenuBar stuff...
