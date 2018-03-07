package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/hpcloud/tail"
	"golang.org/x/net/websocket"
)

var LogFile string
var Listen string

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.StringVar(&LogFile, "log", "/var/log/nginx/access.log", "log file name")
	flag.StringVar(&Listen, "listen", "127.0.0.1:8327", "log file name")
	flag.Parse()

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/follow", websocket.Handler(handleFollow).ServeHTTP)
	http.HandleFunc("/tail", handleTail)
	log.Fatal(http.ListenAndServe(Listen, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t := template.Must(template.New("base").Parse(string(MustAsset("data/index.html"))))
	v := struct {
		Host string
		Log  string
	}{
		r.Host,
		LogFile,
	}
	if err := t.Execute(w, &v); err != nil {
		log.Printf("Template execute failed, err: %v", err)
		return
	}
}

func handleFollow(ws *websocket.Conn) {
	t, err := tail.TailFile(LogFile, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}})
	if err != nil {
		log.Printf("tail file failed, err: %v", err)
		return
	}
	for line := range t.Lines {
		log.Println(line.Text)
		ws.Write([]byte(line.Text))
	}
}

func handleTail(w http.ResponseWriter, r *http.Request) {
	t, err := tail.TailFile(LogFile, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}})
	if err != nil {
		log.Printf("tail file failed, err: %v", err)
		return
	}
	for line := range t.Lines {
		log.Println(line.Text)
		w.Write([]byte(line.Text))
	}
}
