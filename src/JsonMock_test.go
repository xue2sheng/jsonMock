package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Similar to the code under test
var DataDir = "data"
var MockRequestResponseFile = "requestResponseMap.json"

// global due to lazyness
var queryStr string
var dataFile string
var checkUp bool
var gzipOn bool

// To process Json input file
type ReqRes struct {
	Qry string           `json:"query,omitempty"`
	Req *json.RawMessage `json:"req"`
	Res *json.RawMessage `json:"res"`
}

// read extra commandline arguments
func init() {
	flag.StringVar(&queryStr, "queryStr", "http://0.0.0.0/testingEnd?", "Testing End address, including 'debug' parameter if needed")
	mockRequestResponseFile := filepath.Dir(os.Args[0]) + filepath.FromSlash("/") + DataDir + filepath.FromSlash("/") + MockRequestResponseFile
	flag.StringVar(&dataFile, "dataFile", mockRequestResponseFile, "Data File with Request/Response map. No validation will be carried out.")
	flag.BoolVar(&checkUp, "checkUp", true, "Check it out that FastCGI is up and running through a HEAD request.")
	flag.BoolVar(&gzipOn, "gzipOn", false, "Activate GZIP by adding specific header to the request. That might make all tests fail")
	flag.Parse()
}

func TestRequests(t *testing.T) {

	// depends on your NGINX fastcgi configuration
	t.Log("-queryStr=" + queryStr)
	// depends on your test configuration
	t.Log("-dataFile=" + dataFile)
	// depends if the server under test supports HEAD queries
	t.Logf("-checkUp=%t\n", checkUp)

	// call that fastcgi to checkout whether it's up or not
	if checkUp {
		ping, err := http.Head(queryStr)
		if err != nil {
			t.Error("Unable to request for HEAD info to the server.")
			t.Fatal(err)
			t.FailNow()
		}
		if ping.StatusCode != http.StatusOK {
			t.Error("Probably FastCGI down.")
			t.Fatal(ping.Status)
			t.FailNow()
		}
	}

	// grab the real queries to launch
	dataMap, err := ioutil.ReadFile(dataFile)
	if err != nil {
		t.Error("Unable to read Mock Request Response File.")
		t.Fatal(err)
		t.FailNow()
	}

	// process json input
	dec := json.NewDecoder(strings.NewReader(string(dataMap)))
	err = ignoreFirstBracket(dec)
	if err != nil {
		t.Error("Unable to process Mock Request Response File.")
		t.Fatal(err)
		t.FailNow()
	}

	// resquests stats
	failedRequests := 0
	successRequests := 0

	// read object {"req": string, "res": string}
	for dec.More() {

		var rr ReqRes

		err = dec.Decode(&rr)
		if err != nil {
			t.Error("Unable to process Request Response object.")
			continue
		}

		if checkRequest(t, &rr) {
			successRequests++
		} else {
			failedRequests++
		}
	}

	err = ignoreLastBracket(dec)
	if err != nil {
		t.Error("Unable to process Mock Request Response File.")
		t.Fatal(err)
		t.FailNow()
	}

	t.Logf("Failed Requests: %d\n", failedRequests)
	t.Logf("Success Requests: %d\n", successRequests)
	t.Logf("Total requests sent: %d\n", failedRequests+successRequests)

	if failedRequests > 0 {
		t.Errorf("Failed Requests: %d\n", failedRequests)
		t.Fatalf("Failed Requests: %d\n", failedRequests)
		t.FailNow()
	}

}

// process specif request
func checkRequest(t *testing.T, rr *ReqRes) bool {

	query := queryStr
	if len(rr.Qry) > 0 {
		query += rr.Qry
	}

	// create the request
	req, err := toString(rr.Req)
	if err != nil {
		t.Error(err)
		return false
	}
	request, err := http.NewRequest("POST", query, strings.NewReader(req))
	if err != nil {
		t.Error("[" + query + "]" + req + ": " + err.Error())
		return false
	}
	if gzipOn {
		request.Header.Add("Accept-Encoding", "gzip")
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Content-Length", strconv.Itoa(len(req)))

	// making the call
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		t.Error("[" + query + "]" + req + ": " + err.Error())
		return false
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Error("[" + query + "]" + req + ": " + response.Status)
		return false
	}

	// double check the response depending on GZIP usage
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		defer reader.Close()
	default:
		reader = response.Body
	}
	res, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Error("[" + query + "]" + req + ": " + err.Error())
		return false
	}
	responseStr := string(res)

	// what it's read from the file
	expected, err := toString(rr.Res)
	if err != nil {
		t.Error("[" + query + "]" + req + ": " + err.Error())
		return false
	}
	if strings.EqualFold(responseStr, expected) {
		// success
		return true
	}

	// failure
	t.Error("[" + query + "]" + req + ": received->" + responseStr + " expected->" + expected)
	return false
}

// convert into an string
func toString(raw *json.RawMessage) (string, error) {
	if raw != nil {
		noSoRaw, err := json.Marshal(raw)
		if err != nil {
			log.Fatal(err)
			return "", err
		}
		return string(noSoRaw), nil
	} else {
		return "", nil
	}
}

// ignore first bracket when json mock Request Response file is decoded
func ignoreFirstBracket(dec *json.Decoder) error {
	_, err := dec.Token()
	if err != nil {
		log.Fatal(err)
		return errors.New("Unable to process first token at Mock Request Response File")
	}
	return nil
}

// ignore last bracket when json mock Request Response file is decoded
func ignoreLastBracket(dec *json.Decoder) error {
	_, err := dec.Token()
	if err != nil {
		log.Fatal(err)
		return errors.New("Unable to process last token at Mock Request Response File")
	}
	return nil
}

// compact json to make it easy to look into the map for equivalent keys
func compactJson(loose []byte) (string, error) {

	compactedBuffer := new(bytes.Buffer)
	err := json.Compact(compactedBuffer, loose)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return compactedBuffer.String(), nil
}
