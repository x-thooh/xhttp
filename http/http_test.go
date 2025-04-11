package http

import (
	"context"
	"fmt"
	"testing"
)

func TestGet(t *testing.T) {
	for i := 5; i > 0; i-- {
		url := "http://127.0.0.1:8081/json"
		type Ret struct {
			Code int           `json:"code"`
			Msg  string        `json:"msg"`
			Data []interface{} `json:"data"`
		}
		var ret Ret
		if err := Get(context.Background(), url, &ret); err != nil {
			fmt.Println(err)
		}
		fmt.Println(ret)
	}
}

func TestPost(t *testing.T) {
	url := "http://127.0.0.1:8081/json"
	type Ret struct {
		Code int           `json:"code"`
		Msg  interface{}   `json:"msg"`
		Data []interface{} `json:"data"`
	}
	var ret Ret
	if err := Post(context.Background(), url, &ret, WithQuery(map[string]interface{}{
		"a": "A",
	}), WithBody(struct {
		A string `json:"a"`
	}{
		A: "A",
	})); err != nil {
		fmt.Println(err)
	}
	fmt.Println(ret)
}
