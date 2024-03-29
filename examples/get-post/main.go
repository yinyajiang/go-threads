package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/yinyajiang/go-threads"
)

func main() {
	t, err := threads.NewThreads()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	post, err := t.GetPost(3333821180161137169)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var postPretty bytes.Buffer
	if err = json.Indent(&postPretty, post, "", "\t"); err != nil {
		log.Fatal("JSON parse error: ", err)
	}

	fmt.Println(postPretty.String())
}
