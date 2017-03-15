package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

// Request Response map
type RequestResponseMap map[string]string

// helper for HTTP handler queries
type customHandler struct {
	cmux  http.Handler
	rrmap *RequestResponseMap
}

// RequestJsonSchema to validate requests
var RequestJsonSchemaFile = "requestJsonSchema.json"

// ResponseJsonSchema to validate responses
var ResponseJsonSchemaFile = "responseJsonSchema.json"

// MockRequestResponseFile global var due to lazyness
var MockRequestResponseFile = "requestResponseMap.json"

// DebugParameter global var due to lazyness
var DebugParameter = "debug"

func main() {

	host, port, mockRequestResponseFile, requestJsonSchemaFile, responseJsonSchemaFile := cmdLine()
	log.Println("Launched " + host + ":" + port + " MockRequestResponseFile=" + mockRequestResponseFile +
		" RequestJsonSchemaFile=" + requestJsonSchemaFile + " ResponseJsonSchemaFile=" + responseJsonSchemaFile)

	reqresmap, err := validateMockRequestResponseFile(mockRequestResponseFile, requestJsonSchemaFile, responseJsonSchemaFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Number of faked request/response: %d", len(reqresmap))

	mux := mux.NewRouter()
	// bind cmux to mx(route) and rrmap to reqresmap
	fcgiHandler := &customHandler{cmux: mux, rrmap: &reqresmap}
	mux.Path("/").Handler(fcgiHandler)

	listener, _ := net.Listen("tcp", host+":"+port) // see nginx.conf
	if err := fcgi.Serve(listener, fcgiHandler); err != nil {
		log.Fatal(err)
	}
}

// get command line parameters
func cmdLine() (string, string, string, string, string) {

	hostArg := "0.0.0.0"
	portArg := "9797"
	mockRequestResponseFile := filepath.Dir(os.Args[0]) + filepath.FromSlash("/") + MockRequestResponseFile
	requestJsonSchemaFile := filepath.Dir(os.Args[0]) + filepath.FromSlash("/") + RequestJsonSchemaFile
	responseJsonSchemaFile := filepath.Dir(os.Args[0]) + filepath.FromSlash("/") + ResponseJsonSchemaFile

	cmd := strings.Join(os.Args, " ")
	if strings.Contains(cmd, " help") || strings.Contains(cmd, " -help") || strings.Contains(cmd, " --help") ||
		strings.Contains(cmd, " -h") || strings.Contains(cmd, " /?") {
		fmt.Println()
		fmt.Println("Usage: " + os.Args[0] + " <host> <port> <MockRequestResponseFile> <RequestJsonSchema> <ResponseJsonSchema>")
		fmt.Println()
		fmt.Println("host:  Host name for this FastCGI process.   By default " + hostArg)
		fmt.Println("port:  Port number for this FastCGI process. By default " + portArg)
		fmt.Println()
		fmt.Println("MockRequestResponseFile: Fake mapped request/response file. By default " + mockRequestResponseFile)
		fmt.Println("RequestJsonSchemaFile:	  Json Schema to validate requests. By default " + requestJsonSchemaFile)
		fmt.Println("ResponseJsonSchemaFile:  Json Schema to validate responses. By default " + responseJsonSchemaFile)
		fmt.Println()
		fmt.Println("Being a FastCGI, don't forget to properly configure NGINX.")
		fmt.Println()
		os.Exit(0)
	}

	if len(os.Args) > 1 {
		hostArg = os.Args[1]
	}
	if len(os.Args) > 2 {
		portArg = os.Args[2]
	}
	if len(os.Args) > 3 {
		mockRequestResponseFile = os.Args[3]
	}
	if len(os.Args) > 4 {
		requestJsonSchemaFile = os.Args[4]
	}
	if len(os.Args) > 5 {
		responseJsonSchemaFile = os.Args[5]
	}
	return hostArg, portArg, mockRequestResponseFile, requestJsonSchemaFile, responseJsonSchemaFile
}

// validate fake request response map against their json schemas
func validateMockRequestResponseFile(mockRequestResponseFile string, requestJsonSchemaFile string, responseJsonSchemaFile string) (RequestResponseMap, error) {
	var err error
	var reqresmap RequestResponseMap = make(map[string]string)

	mock, err := validateMockInput(mockRequestResponseFile)
	if err != nil {
		return reqresmap, err
	}

	req, err := ioutil.ReadFile(requestJsonSchemaFile)
	if err != nil {
		log.Fatal(err)
		return reqresmap, errors.New("Unable to read Request Json Schema File.")
	}

	res, err := ioutil.ReadFile(responseJsonSchemaFile)
	if err != nil {
		log.Fatal(err)
		return reqresmap, errors.New("Unable to read Response Json Schema File.")
	}

	reqJsonSchema := gojsonschema.NewStringLoader(string(req))
	resJsonSchema := gojsonschema.NewStringLoader(string(res))

	type ReqRes struct {
		Req, Res string
	}
	dec := json.NewDecoder(strings.NewReader(string(mock)))

	err = ignoreFirstBracket(dec)
	if err != nil {
		return reqresmap, err
	}

	// read object {"req": string, "res": string}
	for dec.More() {
		var rr ReqRes
		err = dec.Decode(&rr)
		if err != nil {
			log.Fatal(err)
			return reqresmap, errors.New("Unable to process object at Mock Request Response File")
		}
		fmt.Printf("%v -> %v\n", rr.Req, rr.Res)

		if !validateRequest(reqJsonSchema, rr.Req) {
			continue
		}

		if !validateResponse(resJsonSchema, rr.Res) {
			continue
		}

		// add pair to the map but after compacting those json
		reqBuffer := new(bytes.Buffer)
		err = json.Compact(reqBuffer, []byte(rr.Req))
		if err != nil {
			log.Fatal(err)
			log.Println("This request will be ignored")
			continue
		}
		resBuffer := new(bytes.Buffer)
		err = json.Compact(resBuffer, []byte(rr.Res))
		if err != nil {
			log.Fatal(err)
			log.Println("That response will be ignored")
			continue
		}
		key := reqBuffer.String()
		value := resBuffer.String()
		reqresmap[key] = value

	}

	err = ignoreLastBracket(dec)
	if err != nil {
		return reqresmap, err
	}

	// return result
	if len(reqresmap) == 0 {
		err = errors.New("Unable to validate any entry at Mock Request Response File")
	}
	return reqresmap, err
}

// validation request
func validateRequest(reqJsonSchema gojsonschema.JSONLoader, rrReq string) bool {

	result, err := gojsonschema.Validate(reqJsonSchema, gojsonschema.NewStringLoader(rrReq))
	if err != nil {
		log.Fatal(err)
		log.Println("This request will be ignored")
		return false
	}
	if !result.Valid() {
		log.Println("Request is not valid. See errors: ")
		for _, desc := range result.Errors() {
			log.Printf("- %s\n", desc)
		}
		log.Println("That request will be ignored")
		return false
	}

	return true
}

// validation response
func validateResponse(resJsonSchema gojsonschema.JSONLoader, rrRes string) bool {

	result, err := gojsonschema.Validate(resJsonSchema, gojsonschema.NewStringLoader(rrRes))
	if err != nil {
		log.Fatal(err)
		log.Println("This response will be ignored")
		return false
	}
	if !result.Valid() {
		log.Println("Response is not valid. See errors: ")
		for _, desc := range result.Errors() {
			log.Printf("- %s\n", desc)
		}
		log.Println("That response will be ignored")
		return false
	}

	return true
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

// validate just mock input
func validateMockInput(mockRequestResponseFile string) ([]byte, error) {

	mock, err := ioutil.ReadFile(mockRequestResponseFile)
	if err != nil {
		log.Fatal(err)
		return mock, errors.New("Unable to read Mock Request Response File.")
	}

	// validate the own mock input
	mockJsonSchema := gojsonschema.NewStringLoader(`{ 
		"$schema": "http://json-schema.org/draft-04/schema#",
  		"title": "Mock Request Response Json Schema",
  		"description": "version 0.0.1",
    	"type": "array",
    	"items": {
    		"type": "object",
    		"properties": {
      			"req": {
        			"type": "string"
      			},
      			"res": {
        			"type": "string"
      			}
    		},
    		"required": [
      			"req",
      			"res"
    		]
  		}
	}`)

	result, err := gojsonschema.Validate(mockJsonSchema, gojsonschema.NewStringLoader(string(mock)))
	if err != nil {
		log.Fatal(err)
		return mock, errors.New("Unable to process mock Json Schema")
	}
	if !result.Valid() {
		log.Println("Mock Request Response File is not valid. See errors: ")
		for _, desc := range result.Errors() {
			log.Printf("- %s\n", desc)
		}
		return mock, errors.New("Invalid Mock Request Response File")
	}

	// success
	return mock, nil
}

// must have at least ServeHTTP(), otherwise you will get this error
// *customHandler does not implement http.Handler (missing ServeHTTP method)
func (c *customHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	debug := (r.URL.Query()[DebugParameter] != nil)

	message := "{ \"id\": 1 }"
	w.Header().Set("Content-Lenghth", strconv.Itoa(len(message)))
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(message)); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		if debug {
			log.Println(err)
		}
	}

	/*
		//parameter := r.URL.Query().Get(................)
		//if decrypted, err := decrypt(key, parameter, debug); err != nil {
		if debug {
			//http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			if debug {
				//log.Println(err)
			}
			//} else if err = grabImage(decrypted, w, debug); err != nil {
		} else {
			//http.Error(w, err.Error(), http.StatusConflict)
			if debug {
				//log.Println(err)
			}
		}
	*/
}
