# go-lazyhttp
go-lazyhttp is an http wrapper around the standard go http library keeping you in full control of the underlying objects but providing convenience functions like pre request hooks, authentication hooks, post request hooks, extendable decoders, authentication and more. All while staying true to go's standard implementations. 

It is a project aiming for a simple to use http client wrapper that is not hiding anything about it's behaviour behind opinionated structure. It is not trying to improve buffer handling, dialing, transport, connection establishment etc. Its purpose is only to reduce writing boilerplate code and providing a stable approach for everyday http client implementation.

## Usage
This project is currently in experimental stage and should not be used in production projects.

## Design
Take a look into DESIGN.md for further design goals and the next TODOs.

## Example

### Simple GET request
```go
// use the default go http client and set a timeout
httpClient := http.DefaultClient
httpClient.Timeout = 30 * time.Second

// create a new lazyhttp client
client := lazyhttp.NewClient(
	lazyhttp.WithHost("http://localhost:8080/"),
	lazyhttp.WithHttpClient(httpClient),
)

// create a new request with a given context
req, err := http.NewRequestWithContext(contex.TODO(), http.MethodGet, "/test", nil)
if err != nil {
	log.Errorf("error creating new request: %#v", err)
	return 
}

res, err := client.Do(req)
if err != nil {
	log.Errorf("error making request: %#v", err)
	return
}

type someDataType struct {
	Value string
}

var data someDataType
err = lazyhttp.DecodeJson(res.Body, &data)
if err != nil {
	log.Errorf("error decoding response: %#v", err)
	return
}
```