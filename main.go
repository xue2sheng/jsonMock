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
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

type QueryResponse struct {
	query    string
	response string
}

// Request Response map
type RequestResponseMap map[string]QueryResponse

// helper for HTTP handler queries
type customHandler struct {
	cmux        http.Handler
	rrmap       *RequestResponseMap
	forcedDebug bool
}

// RequestJsonSchema to validate requests
var RequestJsonSchemaFile = "requestJsonSchema.json"

// ResponseJsonSchema to validate responses
var ResponseJsonSchemaFile = "responseJsonSchema.json"

// MockRequestResponseFile global var due to lazyness
var MockRequestResponseFile = "requestResponseMap.json"

// DebugParameter global var due to lazyness
var DebugParameter = "debug"
var ForcedDebug = false

func main() {

	host, port, mockRequestResponseFile, requestJsonSchemaFile, responseJsonSchemaFile, forcedDebug := cmdLine()
	log.Printf("Launched "+host+":"+port+" MockRequestResponseFile="+mockRequestResponseFile+
		" RequestJsonSchemaFile="+requestJsonSchemaFile+" ResponseJsonSchemaFile="+responseJsonSchemaFile+
		" ForcedDebug=%t", forcedDebug)

	reqresmap, err := validateMockRequestResponseFile(mockRequestResponseFile, requestJsonSchemaFile, responseJsonSchemaFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Number of fake request/response: %d", len(reqresmap))

	mux := mux.NewRouter()
	// bind cmux to mx(route) and rrmap to reqresmap
	fcgiHandler := &customHandler{cmux: mux, rrmap: &reqresmap, forcedDebug: forcedDebug}
	mux.Path("/").Handler(fcgiHandler)

	listener, _ := net.Listen("tcp", host+":"+port) // see nginx.conf
	if err := fcgi.Serve(listener, fcgiHandler); err != nil {
		log.Fatal(err)
	}
}

// get command line parameters
func cmdLine() (string, string, string, string, string, bool) {

	hostArg := "0.0.0.0"
	portArg := "9797"
	mockRequestResponseFile := filepath.Dir(os.Args[0]) + filepath.FromSlash("/") + MockRequestResponseFile
	requestJsonSchemaFile := filepath.Dir(os.Args[0]) + filepath.FromSlash("/") + RequestJsonSchemaFile
	responseJsonSchemaFile := filepath.Dir(os.Args[0]) + filepath.FromSlash("/") + ResponseJsonSchemaFile
	forcedDebug := ForcedDebug

	cmd := strings.Join(os.Args, " ")
	if strings.Contains(cmd, " help") || strings.Contains(cmd, " -help") || strings.Contains(cmd, " --help") ||
		strings.Contains(cmd, " -h") || strings.Contains(cmd, " /?") {
		fmt.Println()
		fmt.Println("Usage: " + os.Args[0] + " <host> <port> <MockRequestResponseFile> <RequestJsonSchema> <ResponseJsonSchema> <ForcedDebug>")
		fmt.Println()
		fmt.Println("host:  Host name for this FastCGI process.   By default " + hostArg)
		fmt.Println("port:  Port number for this FastCGI process. By default " + portArg)
		fmt.Println()
		fmt.Println("MockRequestResponseFile: Fake mapped request/response file. By default " + mockRequestResponseFile)
		fmt.Println("RequestJsonSchemaFile:	  Json Schema to validate requests. By default " + requestJsonSchemaFile)
		fmt.Println("ResponseJsonSchemaFile:  Json Schema to validate responses. By default " + responseJsonSchemaFile)
		fmt.Println()
		fmt.Printf("ForcedDebug:  Flag to force debug mode. By default %b\n", forcedDebug)
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
	if len(os.Args) > 6 {
		if os.Args[6] == "ForcedDebug" || os.Args[6] == "forcedDebug" || os.Args[6] == "true" || os.Args[6] == "TRUE" || os.Args[6] == "True" {
			forcedDebug = true
		}
	}
	return hostArg, portArg, mockRequestResponseFile, requestJsonSchemaFile, responseJsonSchemaFile, forcedDebug
}

// validate fake request response map against their json schemas
func validateMockRequestResponseFile(mockRequestResponseFile string, requestJsonSchemaFile string, responseJsonSchemaFile string) (RequestResponseMap, error) {
	var err error
	var reqresmap RequestResponseMap = make(map[string]QueryResponse)

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
		Qry string `json:"query,omitempty"`
		Req string `json:"req"`
		Res string `json:"res"`
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
		if len(rr.Qry) > 0 {
			log.Printf("%v %v -> %v\n", rr.Qry, rr.Req, rr.Res)
		} else {
			log.Printf("%v -> %v\n", rr.Req, rr.Res)
		}

		if !validateRequest(reqJsonSchema, rr.Req) {
			continue
		}

		if !validateResponse(resJsonSchema, rr.Res) {
			continue
		}

		// add pair to the map but after compacting those json
		key, err := compactJson([]byte(rr.Req))
		if err != nil {
			log.Println("This request will be ignored")
			continue
		}
		if len(rr.Qry) > 0 {
			// key must take into account as well the provided query
			key = "[" + rr.Qry + "]" + key
		}
		response, err := compactJson([]byte(rr.Res))
		if err != nil {
			log.Println("That response will be ignored")
			continue
		}
		var value QueryResponse
		value.response = response
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
      		   },
               "query": {
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

	debug := (r.URL.Query()[DebugParameter] != nil) || c.forcedDebug

	// GET params as a string
	query := QueryAsString(r)
	if debug {
		log.Println(query)
	}

	if r.ContentLength > 0 {

		// get body request to process
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			if debug {
				log.Println(err)
			}
		}

		if debug {
			log.Println("Body received: " + string(body))
		}

		// avoid processing before having booted up completely
		if c.rrmap != nil && len(*c.rrmap) > 0 {

			key, err := compactJson(body)
			if err != nil {
				if debug {
					log.Print(err)
				}
			}
			if len(query) > 0 {
				key = "[" + query + "]" + key
			}
			value := (*c.rrmap)[key]
			if len(value.response) > 0 {
				w.Header().Set("Content-Lenghth", strconv.Itoa(len(value.response)))
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(value.response)); err != nil {
					http.Error(w, err.Error(), http.StatusUnprocessableEntity)
					if debug {
						log.Println(err)
					}
				}
				if debug {
					log.Println("Sent back: " + value.response)
				}
			} else {
				http.Error(w, "key not found at internal cache", http.StatusNoContent)
				if debug {
					log.Println("key not found at internal cache")
				}
			}
		}

	} else {
		http.Error(w, "empty request body received", http.StatusNoContent)
		if debug {
			log.Println("empty request body received")
		}
	}

	if debug {
		log.Printf("Processed request of %d bytes", r.ContentLength)
	}
}

// convert query parameter into a string to be used as index in the map
func QueryAsString(r *http.Request) string {

	// try to get IN ORDER all the parameters
	keys := []string{}
	for k, _ := range r.URL.Query() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	query := ""
	for _, k := range keys {
		if k == DebugParameter {
			continue
		}
		if len(query) > 0 {
			query += "&"
		}
		v := r.URL.Query()[k]
		if len(v) > 0 { // there might be repeated params
			query += k + "="
			for i, w := range v {
				if i > 0 {
					query += ","
				}
				query += w
			}
		} else {
			// is a Flag
			query += k
		}
	}
	return query
}
