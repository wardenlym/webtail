package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var LogFile string
var Listen string

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.StringVar(&LogFile, "log", "", "log file name")
	flag.StringVar(&Listen, "listen", "127.0.0.1:8327", "log file name")
	flag.Parse()

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/follow", serveWs)
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	go writer(ws)
}

func writer(ws *websocket.Conn) {
	defer ws.Close()
	var r *bufio.Reader
	if LogFile == "" {
		r = bufio.NewReader(os.Stdin)
	} else {
		f, _ := os.Open(LogFile)

		fi, err := f.Stat()
		if err != nil {
			fmt.Println(err)
			ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		}
		lengh := fi.Size() - int64(256)

		if lengh < 0 {
			lengh = fi.Size()
		}
		fmt.Println(lengh, fi.Size())
		ret, err := f.Seek((0 - lengh), io.SeekEnd)
		fmt.Println(ret, err)
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		}

		r = bufio.NewReader(f)
		defer f.Close()
	}
	for {
		p, err := r.ReadBytes('\n')

		if len(p) != 0 {
			ws.WriteMessage(websocket.TextMessage, p)
		}

		if err != nil {
			if err == io.EOF {
				time.Sleep(1 * time.Second)
			} else {
				ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			}
		}
	}
}

// func handleFollow(ws *websocket.Conn) {

// 	ws.Write([]byte("webtail -f\n"))
// 	file, err := os.Open(LogFile)
// 	if err != nil {
// 		log.Println(err)
// 		ws.Write([]byte(err.Error() + "\n"))
// 	}

// 	fi, err := file.Stat()
// 	if err != nil {
// 		log.Println(err)
// 		ws.Write([]byte(err.Error() + "\n"))
// 	}

// 	buf := make([]byte, 256)
// 	l := fi.Size() - int64(len(buf))
// 	log.Println(l, fi.Size(), int64(len(buf)))
// 	if l < 0 {
// 		l = 0
// 	}
// 	n, err := file.ReadAt(buf, l)
// 	if err != nil {
// 		log.Println(err)
// 		ws.Write([]byte(err.Error() + "\n"))
// 	}
// 	buf = buf[:n]
// 	log.Println(string(buf))

// 	ss := strings.Split(string(buf), "\n")
// 	for _, s := range ss {
// 		ws.Write([]byte(s))
// 	}

// 	file.Close()

// 	t, err := tail.TailFile(LogFile, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}})
// 	if err != nil {
// 		log.Printf("tail file failed, err: %v", err)
// 		return
// 	}
// 	for line := range t.Lines {
// 		log.Println(line.Text)
// 		ws.Write([]byte(line.Text))
// 	}
// }
