# go-lazyhttp
Go lazy http is an http wrapper around the standard go http library keeping you in full control of the underlying objects but providing convenience functions like pre request hooks, authentication hooks, post request hooks, extendable decoders and more. All while staying true to go's standard implementations. 

## Usage
This project is currently in experimental stage and should not be used in production projects.

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
	req, err := lazyhttp.NewRequestWithContext(contex.TODO(), http.MethodGet, "/test")
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