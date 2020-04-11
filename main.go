package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type Config struct {
	root  string
	files []string
	Date  string
}

func main() {
	config := &Config{}
	config.root = "./images"

	err := filepath.Walk(config.root, visit(&config.files))
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
	sfiles := make([]string, len(config.files))
	copy(sfiles, config.files)

	t := time.Now()
	today := t.Format("20060102")
	date, err := strconv.ParseInt(today, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(date)
	rand.Shuffle(len(sfiles), func(i, j int) { sfiles[i], sfiles[j] = sfiles[j], sfiles[i] })
	return sfiles
}

func (config *Config) ConnWs(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 15000000, 1024)
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

		files := shuffleFiles(config)

		res["a"] = "a"
		log.Println(res)

		timings := []int{15, 15, 15, 15, 15, 15, 30, 30, 30, 30, 45, 45, 60, 60, 300, 300, 300, 300, 300}
		for i, s := range timings {
			log.Println(files[i])
			res["img"] = files[i]
			res["time"] = s
			res["index"] = i + 1
			res["indexLength"] = len(timings)

			time.Sleep(2 * time.Second)
			if err = ws.WriteJSON(&res); err != nil {
				fmt.Println("watch dir - Write : " + err.Error())
				return
			}

			time.Sleep(time.Duration(s) * time.Second)
		}
	}
}
