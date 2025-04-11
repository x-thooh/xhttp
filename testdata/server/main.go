package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	nh "net/http"
	"time"
)

func main() {
	nh.HandleFunc("/json", func(w nh.ResponseWriter, r *nh.Request) {
		fmt.Println("request :", r.RemoteAddr, r.Header)
		w.Write([]byte(fmt.Sprintf(`{"code": 1, "msg": %s, "data": [3] }`, func() string {
			bs, _ := ioutil.ReadAll(r.Body)
			return bytes.NewBuffer(bs).String()
		}())))
	})

	nh.HandleFunc("/string", func(w nh.ResponseWriter, r *nh.Request) {
		w.Write([]byte(`"hello world"`))
	})

	log.Fatal((&nh.Server{
		Addr:        ":8081",
		IdleTimeout: 5 * time.Second,
	}).ListenAndServe())
}
