package main

import (
	"fmt"
	"net/http"
	"os"
	"testing"
)

func TestRequests(t *testing.T) {

	// depends on your NGINX fastcgi configuration
	curlStr := "http://0.0.0.0/testingEnd?debug"
	if len(os.Args) > 1 {
		curlStr = os.Args[1]
	}
	t.Log(curlStr)
	fmt.Println(curlStr)

	// call that fastcgi to checkout whether it's up or not
	ping, err := http.Head(curlStr)
	if err != nil {
		t.Error("Unable to request for HEAD info to the server.")
		t.Fatal(err)
		t.FailNow()
	}
	if ping.StatusCode != http.StatusOK {
		fmt.Println(ping.Status)
		t.Error("Probably FastCGI down.")
		t.Fatal(ping.Status)
		t.FailNow()
	}

	/*
		// grab the real queries to launch
		rrMap, err := ioutil.ReadFile(dataFile)
		if err != nil {
			t.Error("Unable to read Mock Request Response File.")
			t.Fatal(err)
			t.FailNow()
		}
		t.Logf("%d requests to try out\n", len(rrMap))
		fmt.Printf("%d requests to try out\n", len(rrMap))
	*/
	/*
		if response.ContentLength > 0 {
			var responseInfo []byte
			response.Body.Read(responseInfo)
			t.Log(string(responseInfo))
			fmt.Println(string(responseInfo))
		} else {
			t.Error("Empty Response Body")
			t.FailNow()
		}
	*/
}
