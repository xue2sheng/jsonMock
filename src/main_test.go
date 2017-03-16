package main

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

func TestRequests(t *testing.T) {

	curlStr := "http://0.0.0.0/testingEnd?debug"
	if len(os.Args) > 1 {
		curlStr = os.Args[1]
	}
	t.Log(curlStr)
	fmt.Println(curlStr)

	// call that fastcgi
	response, err := http.Get(curlStr)
	if err != nil {
		t.Error(err)
		return
	}
	defer response.Body.Close()
	var responseInfo []byte
	response.Body.Read(responseInfo)

	t.Log(string(responseInfo))
	fmt.Println(string(responseInfo))
}
