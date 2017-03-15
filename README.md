# Json Mock

Simple FastCGI golang implemented Json Mock in order to 'fake' validated json requests/responses.

![Mock Server](/images/mockServer.jpeg)

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


