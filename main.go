package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/koding/multiconfig"
)

type Config struct {
	Root  string `default:"./images/"`
	Seq   string `default:"15,15,15,15,15,15,30,30,30,30,45,45,60,60,300,300,300,300,300"`
	Files []string
}

func main() {
	config := &Config{}
	multiconfig.New().MustLoad(config)

	err := filepath.Walk(config.Root, visit(&config.Files))
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/connws/", config.ConnWs)
	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func visit(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		if info.IsDir() && info.Name() == "@eaDir" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		*files = append(*files, path)
		return nil
	}
}

func shuffleFiles(config *Config) []string {
	f := make([]string, len(config.Files))
	copy(f, config.Files)

	t := time.Now()
	day := t.Format("20060102")
	date, err := strconv.ParseInt(day, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(date)
	rand.Shuffle(len(f), func(i, j int) { f[i], f[j] = f[j], f[i] })
	return f
}

func (config *Config) ConnWs(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}

	res := map[string]interface{}{}
	for {
		if err = ws.ReadJSON(&res); err != nil {
			if err.Error() == "EOF" {
				return
			}
			// ErrShortWrite means a write accepted fewer bytes than requested then failed to return an explicit error.
			if err.Error() == "unexpected EOF" {
				return
			}
			fmt.Println("Read : " + err.Error())
			return
		}

		f := shuffleFiles(config)
		seq := strings.Split(config.Seq, ",")

		res["a"] = "a"
		log.Println(res)

		for i, s := range seq {
			log.Println(f[i])
			res["img"] = f[i]
			res["time"] = s
			res["index"] = i + 1
			res["total"] = len(seq)

			time.Sleep(2 * time.Second)
			if err = ws.WriteJSON(&res); err != nil {
				fmt.Println("watch dir - Write : " + err.Error())
				return
			}

			s, _ := strconv.Atoi(s)

			time.Sleep(time.Duration(s)*time.Second + 2)
		}

		// graceful close connection to client
		if err = ws.Close(); err != nil {
			fmt.Println("close error: " + err.Error())
			return
		}
	}
}
