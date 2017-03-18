# Json Mock

Simple FastCGI golang implemented Json Mock in order to 'fake' validated json requests/responses.

![Mock Server](/images/mockServer.jpeg)

## Dependencies

Some *golang 3rd party libraries* have been used:

    go get github.com/gorilla/mux
	go get github.com/xeipuuv/gojsonschema
    
[gorilla/mux](http://www.gorillatoolkit.org/pkg/mux) by [Diego Siquiera](https://github.com/DiSiqueira) makes it easier to serve *FastCGI* requests and [xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema) by [xeipuuv](https://github.com/xeipuuv/gojsonschema) simpilfies *json schema* validations.

## Getting a simple mock server to simulate client's behaviour

### Curl queries

The query can be simulated using **curl**. For example, a typical call might be:

    curl -vvv -H 'Content-Type: application/json' -H 'Accept-Encoding: gzip' "http://0.0.0.0/testingEnd" -d '{"test": 1, "id": "1"}'


### NGINX configuration

Being a FastCGI that processes request body and probably responses with a **gzipped** json, don't forget:

#### GZIP

Nginx configuration for GZIP WITH THE CORRECT ERROR CODE (200) in the response:

     gzip on;
     gzip_vary on;
     gzip_proxied any;
     gzip_comp_level 6;
     gzip_buffers 16 8k;
     gzip_http_version 1.1;

     gzip_types text/plain text/css application/json application/javascript application/x-javascript text/javascript text/xml application/xml application/rss+xml application/atom+xml application/rdf+xml;

 
#### POST request body

Nginx configuration for passing the POST body:

     fastcgi_param  REQUEST_BODY       $request_body;

### Windows 10

You'd better install **ninja build** command and added to *path enviroment variable*. This way you can build:

     mkdir build
     cd build
     cmake .. -G "Ninja" -DJsonMock_TEST
     ninja
     ninja JsonMock.test
     
