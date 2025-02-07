package main

import (
	"fmt"
	"log"
	"mygeecache"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	mygeecache.NewGroup("scores", 2<<10, mygeecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[slowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	addr := "localhost:9999"
	peers := mygeecache.NewHTTPPool(addr)
	log.Println("mygeecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
