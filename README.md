# Json Mock

Simple FastCGI golang implemented Json Mock in order to 'fake' validated json requests/responses.

![Mock Server](/images/mockServer.jpeg)

Although the main development environment was **Linux** (debian/opensuse), some tips will be provided for your **Windows** and **Apple** boxes. Being a testing tool, we'd better try to be able to work on different operating systems.

## Dependencies

Some *golang 3rd party libraries* have been used:

    go get github.com/gorilla/mux
    go get github.com/xeipuuv/gojsonschema
    
[gorilla/mux](http://www.gorillatoolkit.org/pkg/mux) by [Diego Siqueira](https://github.com/DiSiqueira) makes it easier to serve *FastCGI* requests and [xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema) by [xeipuuv](https://github.com/xeipuuv/gojsonschema) simpilfies *json schema* validations.

## Getting a simple mock server to simulate client's behaviour

### Curl queries

The query can be simulated using **curl**. For example, a typical call might be:

    curl -vvv -H 'Content-Type: application/json' -H 'Accept-Encoding: gzip' "http://0.0.0.0/testingEnd" -d '{"test": 1, "id": "1"}'

### NGINX configuration

On **macOS** systems, it's highly likely that your *nginx* runs on a different port due to security reasons. For example, you might be ending at **8080** port. On that case, you could redirect that port to the usual **80** in oder to avoid having to invoke your test cases with *http://0.0.0.0:8080/testingEnd*.

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

## Windows 10

By installing **ninja build** command and add it to the *path environment varible* at your *Powershell console*, you might spare yourself some pain in the neck to prevent usual Windows *make* warnings/errors:

     mkdir build
     cd build
     cmake .. -G "Ninja" -DJsonMock_TEST=1
     ninja
     ninja JsonMock.test

**NGINX** on Windows 10 won't be so fast as on Linux but the configuration it's very similar. Again like **macOS**, you'd better choose port *8080* for your local nginx.

Regarding to the commandline **curl.exe** (to avoid Powershell *curl* alias), take into account to escape properly all *quotation* marks in the body message:

     curl.exe -vvv -H 'Content-Type: application/json' -H 'Accept-Encoding: gzip' "http://localhost:8080/testingEnd" -d '{\"test\": 1, \"id\": \"1\"}'     
