package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	tr := &http.Transport{
		DisableKeepAlives: false,
	}
	c := &http.Client{
		Transport: tr,
	}
	req, err := http.NewRequest("Get", "http://127.0.0.1:8081/json", nil)
	if err != nil {
		panic(err)
	}

	// req.Header.Set("Connection", "close")
	for i := 0; i < 5; i++ {
		resp, err := c.Do(req)
		if err != nil {
			fmt.Println("http get error:", err)
			return
		}
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("read body error:", err)
			return
		}
		log.Println("response body:", string(b))
		time.Sleep(6 * time.Second)
	}

}
