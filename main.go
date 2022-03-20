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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const APP_NAME = "Notification App"

func main() {
	window, state := configureApp()
	subscribeToServer()
	go listenForUpdates(state)
	window.ShowAndRun()
}

//var MYADDR = "127.0.0.1:9090"
//var SERVERADDR = "http://127.0.0.1:8080"
var SERVERADDR = "http://localhost:8080"
var MYADDR = "172.23.128.1:9090"

func subscribeToServer() {
	postBody, _ := json.Marshal(map[string]string{
		"myaddress": MYADDR,
	})
	responseBody := bytes.NewBuffer(postBody)
	fmt.Printf("connecting to %s\n", SERVERADDR)
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

func listenForUpdates(state *AppState) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// storeIP(w, r, state)
			readAndUpdate(r, state)
			w.WriteHeader(200)
		} else if r.Method == "GET" {
			fmt.Println("received a get!")
			w.WriteHeader(200)
		} else {
			w.WriteHeader(404)
			// only post requests are allowed
		}
	})
	fmt.Println("listening for response on 9090 ")
	http.ListenAndServe(":9090", nil)
}

func readAndUpdate(r *http.Request, state *AppState) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var urls []string
	err = json.Unmarshal(body, &urls)
	if err != nil {
		log.Fatal(err)
	}

	changeFound := false

	for _, urlString := range urls {
		if _, exists := (*state).SeenUrls[urlString]; !exists {
			state.SeenUrls[urlString] = false
			changeFound = true
		}
	}

	if changeFound {
		state.EmitNotification("new post found", "")
		state.UpdateVisibleUrls()
	}
}

type AppState struct {
	SeenUrls        map[string]bool
	VisibleUrls     []string
	notifyCallback  func(string, string)
	refreshCallback func()
}

func (s *AppState) UpdateVisibleUrls() {
	keys := []string{}
	for url, seen := range s.SeenUrls {
		if !seen {
			keys = append(keys, url)
		}
	}
	s.VisibleUrls = keys
	s.refreshCallback()
}

func (s *AppState) EmitNotification(title string, body string) {
	s.notifyCallback(title, body)
}

func (s *AppState) SetRefreshCallback(cb func()) {
	s.refreshCallback = cb
}

func (s *AppState) SetNotifyCallback(cb func(string, string)) {
	s.notifyCallback = cb
}

func configureApp() (fyne.Window, *AppState) {
	a := app.New()
	w := a.NewWindow(APP_NAME)

	state := AppState{
		SeenUrls:    make(map[string]bool),
		VisibleUrls: make([]string, 0, 0),
	}

	var list *widget.List
	list = widget.NewList(func() int {
		return len(state.VisibleUrls)
	},
		func() fyne.CanvasObject {
			label := widget.NewLabel("default label")
			open_button := widget.NewButton("Open", nil)
			clear_button := widget.NewButton("Clear", nil)
			group := container.New(layout.NewHBoxLayout(), label, open_button, clear_button)

			open_button.OnTapped = func() {
				u, e := url.ParseRequestURI(label.Text)
				if e == nil {
					e = a.OpenURL(u)
					if e != nil {
						fmt.Printf("error opening url: %s", e.Error())
					}
				}
			}

			clear_button.OnTapped = func() {
				state.SeenUrls[label.Text] = true
				state.UpdateVisibleUrls()
			}

			return group
		},
		func(itemId int, o fyne.CanvasObject) {
			labelObj := o.(*fyne.Container).Objects[0]
			labelText := state.VisibleUrls[itemId]
			labelObj.(*widget.Label).SetText(labelText)
		},
	)

	state.SetNotifyCallback(func(s1, s2 string) {
		a.SendNotification(fyne.NewNotification(s1, s2))
	})
	state.SetRefreshCallback(list.Refresh)

	w.SetContent(container.NewBorder(nil, nil, nil, nil, list))

	return w, &state
}
