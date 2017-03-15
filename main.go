package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

// helper for HTTP handler queries
type customHandler struct {
	cmux http.Handler
}

// Request Response map
type RequestResponseMap map[string]string

// RequestJsonSchema to validate requests
var RequestJsonSchemaFile = "requestJsonSchema.json"

// ResponseJsonSchema to validate responses
var ResponseJsonSchemaFile = "responseJsonSchema.json"

// MockRequestResponseFile global var due to lazyness
var MockRequestResponseFile = "requestResponseMap.json"

// DebugParameter global var due to lazyness
var DebugParameter = "debug"

func main() {
	mux := mux.NewRouter()
	// bind cmux to mx(route)
	fcgiHandler := &customHandler{cmux: mux}
	mux.Path("/").Handler(fcgiHandler)

	host, port, mockRequestResponseFile, requestJsonSchemaFile, responseJsonSchemaFile := cmdLine()
	log.Println("Launched " + host + ":" + port + " MockRequestResponseFile=" + mockRequestResponseFile +
		" RequestJsonSchemaFile=" + requestJsonSchemaFile + " ResponseJsonSchemaFile=" + responseJsonSchemaFile)

	reqresmap, err := validateMockRequestResponseFile(mockRequestResponseFile, requestJsonSchemaFile, responseJsonSchemaFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Number of faked request/respnse: %d", len(reqresmap))

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
	var reqresmap RequestResponseMap

	mock, err := ioutil.ReadFile(mockRequestResponseFile)
	if err != nil {
		return reqresmap, errors.New("Unable to read Mock Request Response File.")
	}

	req, err := ioutil.ReadFile(requestJsonSchemaFile)
	if err != nil {
		return reqresmap, errors.New("Unable to read Request Json Schema File.")
	}

	res, err := ioutil.ReadFile(responseJsonSchemaFile)
	if err != nil {
		return reqresmap, errors.New("Unable to read Response Json Schema File.")
	}

	// validate the own mock input
	mockJsonSchema := gojsonschema.NewStringLoader(`{ "$schema": "http://json-schema.org/draft-04/schema#", "title": "Mock Request Response Json Schema", "description": "version 0.0.1", "type": "object", "properties": { "map": { "type": "array", "items": { "type": "object", "properties": { "req": { "type": "object" }, "res": { "type": "object" } }, "required": [ "req","res" ] } } }, "required": [ "map" ] }`)
	mockJson := gojsonschema.NewStringLoader(string(mock))
	result, err := gojsonschema.Validate(mockJsonSchema, mockJson)
	if err != nil {
		return reqresmap, errors.New("Unable to process mock Json Schema")
	}
	if !result.Valid() {
		log.Println("Mock Request Response File is not valid. See errors: ")
		for _, desc := range result.Errors() {
			log.Printf("- %s\n", desc)
		}
		return reqresmap, errors.New("Invalid Mock Request Response File")
	}

	fmt.Println(string(req))
	fmt.Println(string(res))

	if nil == reqresmap {
		err = errors.New("Unable to validate Mock Request Response File")
	}
	return reqresmap, err
}

// must have at least ServeHTTP(), otherwise you will get this error
// *customHandler does not implement http.Handler (missing ServeHTTP method)
func (c *customHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//parameter := r.URL.Query().Get(EncryptedPathParameter)
	debug := (r.URL.Query()[DebugParameter] != nil)
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
}
