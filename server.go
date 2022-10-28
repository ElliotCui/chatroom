package main

import (
	"html/template"
	"log"
	"net/http"
)

type Room struct {
	Id       string
	Room     string
	UserName string
}

func main() {
	engine := newWsEngine()
	go engine.launch()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] info: homepage")
	})
	http.HandleFunc("/client", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		roomName := params.Get("room")
		userName := params.Get("user")
		room := &Room{
			Id:       roomName,
			Room:     roomName,
			UserName: userName,
		}
		tmpl := template.Must(template.ParseFiles("./views/test/client.html"))
		tmpl.Execute(w, room)
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] info: ws链接中")

		serveWs(engine, w, r)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("[ERROR] ListenAndServe: ", err)
	}
}
