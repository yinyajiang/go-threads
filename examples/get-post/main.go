package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/yinyajiang/go-threads"
)

func getPostIDfromURL(postURL string) string {
	if postURL == "" {
		return ""
	}
	threadID := strings.Split(postURL, "?")[0]
	if strings.HasSuffix(threadID, "/") {
		threadID = threadID[:len(threadID)-1]
	}
	parts := strings.Split(threadID, "/")
	threadID = parts[len(parts)-1]

	return getPostIDfromThreadID(threadID)
}

func getPostIDfromThreadID(url string) string {

	threadID = strings.Split(threadID, "?")[0]
	threadID = strings.ReplaceAll(threadID, " ", "")
	threadID = strings.ReplaceAll(threadID, "/", "")

	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	postID := "0"

	for _, letter := range threadID {
		index := strings.Index(alphabet, string(letter))
		if index != -1 {
			postID = multiply(postID, "64")
			postID = add(postID, alphabet[index:index+1])
		}
	}

	return postID
}

func main() {
	t, err := threads.NewThreads()
	if err != nil {
		fmt.Println("Error:", err)

		return
	}

	post, err := t.GetPost(3141002295235099165)
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
