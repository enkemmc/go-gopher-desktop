package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

func main() {

	window, data := configureApp()
	subscribeToServer()
	go listenForUpdates(&data)
	// go startPollLoop(&data)
	window.ShowAndRun()
}

var MYADDR = "127.0.0.1:9090"
var SERVERADDR = "http://127.0.0.1:8080"

func subscribeToServer() {
	postBody, _ := json.Marshal(map[string]string{
		"myaddress": MYADDR,
	})
	responseBody := bytes.NewBuffer(postBody)
	fmt.Printf("posting to %s\n", SERVERADDR)
	resp, err := http.Post(SERVERADDR, "application/json", responseBody)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("status code: %d", resp.StatusCode)
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func listenForUpdates(data *binding.StringList) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// storeIP(w, r, state)
			readAndUpdate(r, data)
			w.WriteHeader(200)
		} else {
			w.WriteHeader(404)
			// only post requests are allowed
		}
	})
	fmt.Println("listening for response on 9090 ")
	http.ListenAndServe(":9090", nil)
}

func readAndUpdate(r *http.Request, data *binding.StringList) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var urls []string
	err = json.Unmarshal(body, &urls)
	if err != nil {
		log.Fatal(err)
	}

	newUrls := []string{}
	for _, urlString := range urls {
		found := false
	inner:
		for i := 0; i < (*data).Length(); i++ {
			value, _ := (*data).GetValue(i)
			if value == urlString {
				found = true
				break inner
			}
		}

		if !found {
			newUrls = append(newUrls, urlString)
		}
	}

	for _, urlString := range newUrls {
		err := (*data).Append(urlString)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func configureApp() (fyne.Window, binding.StringList) {
	a := app.New()
	w := a.NewWindow("Hello")

	boundList := binding.NewStringList()
	list := widget.NewListWithData(
		boundList,
		func() fyne.CanvasObject {
			button := widget.NewButton("template", nil)
			button.OnTapped = func() {
				u, e := url.ParseRequestURI(button.Text)
				if e == nil {
					fmt.Println("trying to open url")
					e = a.OpenURL(u)
					if e != nil {
						fmt.Printf("error opening url: %s", e.Error())
					}
				}
				button.Hide()
			}
			return button
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			str, _ := i.(binding.String).Get()
			o.(*widget.Button).SetText(str)
		},
	)

	w.SetContent(container.NewBorder(nil, nil, nil, nil, list))

	return w, boundList
}

// func startApp(data *[]string) {
// 	a := app.New()
// 	w := a.NewWindow("My App")

// 	boundList := binding.BindStringList(data)

// 	list := widget.NewListWithData(boundList, createItemm, updateItemm)

// 	lenLabel := widget.NewLabel("1")
// 	lenButton := widget.NewButton("update", func() {
// 		boundList.Reload()
// 		labelStr := fmt.Sprint(len(*data))
// 		fmt.Printf("updating label to %s\n", labelStr)
// 		lenLabel.SetText(labelStr)
// 	})

// 	// w.SetContent(container.NewVBox(
// 	// 	lenLabel,
// 	// 	lenButton,
// 	// 	list,
// 	// ))
// 	w.SetContent(container.NewBorder(nil, nil, nil, nil, list, lenLabel, lenButton))

// 	w.ShowAndRun()
// }

// func createItemm() fyne.CanvasObject {
// 	label := widget.NewLabel("template")
// 	return label
// }

// func updateItemm(i binding.DataItem, o fyne.CanvasObject) {
// 	o.(*widget.Label).Bind(i.(binding.String))
// }
