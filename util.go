package threads

import "strings"

func GetPostIDfromURL(u string) int64 {
	path := strings.TrimSuffix(strings.Split(u, "?")[0], "/")
	parts := strings.Split(path, "/")
	u = parts[len(parts)-1]
	return GetPostIDfromThreadID(u)
}

func GetPostIDfromThreadID(threadID string) int64 {
	threadID = strings.Split(threadID, "?")[0]
	threadID = strings.ReplaceAll(threadID, " ", "")
	threadID = strings.ReplaceAll(threadID, "\t", "")
	threadID = strings.ReplaceAll(threadID, "/", "")

	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	var postID int64
	for _, letter := range threadID {
		index := strings.Index(alphabet, string(letter))
		postID = postID*64 + int64(index)
	}
	return postID
}
