package http

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/taskcluster/httpbackoff"
	"github.com/tent/hawk-go"
	D "github.com/tj/go-debug"
)

var (
	// Used for logging based on DEBUG environment variable
	// See github.com/tj/go-debug
	debug = D.Debug("http")
)

// The entry point into all the functionality in this package is to create a
// ConnectionData object. It contains authentication credentials, and a service
// endpoint, which are required for all HTTP operations.
type ConnectionData struct {
	// Client ID required by Hawk
	ClientId string
	// Access Token required by Hawk
	AccessToken string
	// The URL of the API endpoint to hit.
	// Use "https://auth.taskcluster.net/v1" for production.
	// Please note calling auth.New(clientId string, accessToken string) is an
	// alternative way to create an Auth object with BaseURL set to production.
	BaseURL string
	// Whether authentication is enabled (e.g. set to 'false' when using taskcluster-proxy)
	// Please note calling auth.New(clientId string, accessToken string) is an
	// alternative way to create an Auth object with Authenticate set to true.
	Authenticate bool
	// Certificate for temporary credentials
	Certificate string
}

// CallSummary provides information about the underlying http request and
// response issued for a given API call.
type CallSummary struct {
	HttpRequest *http.Request
	// Keep a copy of request body in addition to the *http.Request, since
	// accessing the Body via the *http.Request object, you get a io.ReadCloser
	// - and after the request has been made, the body will have been read, and
	// the data lost... This way, it is still available after the api call
	// returns.
	HttpRequestBody string
	// The Go Type which is marshaled into json and used as the http request
	// body.
	HttpRequestObject interface{}
	HttpResponse      *http.Response
	// Keep a copy of response body in addition to the *http.Response, since
	// accessing the Body via the *http.Response object, you get a
	// io.ReadCloser - and after the response has been read once (to unmarshal
	// json into native go types) the data is lost... This way, it is still
	// available after the api call returns.
	HttpResponseBody string
	// Keep a record of how many http requests were attempted
	Attempts int
}

// APICall is the generic REST API calling method which performs all REST API
// calls for this library.  Each auto-generated REST API method simply is a
// wrapper around this method, calling it with specific specific arguments.
func (connectionData *ConnectionData) APICall(payload interface{}, method, route string, result interface{}, query url.Values) (interface{}, *CallSummary, error) {
	callSummary := new(CallSummary)
	callSummary.HttpRequestObject = payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return result, callSummary, err
	}
	callSummary.HttpRequestBody = string(jsonPayload)

	httpClient := &http.Client{}

	// function to perform http request - we call this using backoff library to
	// have exponential backoff in case of intermittent failures (e.g. network
	// blips or HTTP 5xx errors)
	httpCall := func() (*http.Response, error, error) {
		var ioReader io.Reader = nil
		if reflect.ValueOf(payload).IsValid() && !reflect.ValueOf(payload).IsNil() {
			ioReader = bytes.NewReader(jsonPayload)
		}
		u, err := url.Parse(connectionData.BaseURL + route)
		if err != nil {
			return nil, nil, fmt.Errorf("apiCall url cannot be parsed: '%v', is your BaseURL (%v) set correctly?\n%v\n", connectionData.BaseURL+route, connectionData.BaseURL, err)
		}
		if query != nil {
			u.RawQuery = query.Encode()
		}
		httpRequest, err := http.NewRequest(method, u.String(), ioReader)
		if err != nil {
			return nil, nil, fmt.Errorf("Internal error: apiCall url cannot be parsed although thought to be valid: '%v', is the BaseURL (%v) set correctly?\n%v\n", u.String(), connectionData.BaseURL, err)
		}
		httpRequest.Header.Set("Content-Type", "application/json")
		callSummary.HttpRequest = httpRequest
		// Refresh Authorization header with each call...
		// Only authenticate if client library user wishes to.
		if connectionData.Authenticate {
			credentials := &hawk.Credentials{
				ID:   connectionData.ClientId,
				Key:  connectionData.AccessToken,
				Hash: sha256.New,
			}
			reqAuth := hawk.NewRequestAuth(httpRequest, credentials, 0)
			if connectionData.Certificate != "" {
				reqAuth.Ext = base64.StdEncoding.EncodeToString([]byte("{\"certificate\":" + connectionData.Certificate + "}"))
			}
			httpRequest.Header.Set("Authorization", reqAuth.RequestHeader())
		}
		debug("Making http request: %v", httpRequest)
		resp, err := httpClient.Do(httpRequest)
		return resp, err, nil
	}

	// Make HTTP API calls using an exponential backoff algorithm...
	callSummary.HttpResponse, callSummary.Attempts, err = httpbackoff.Retry(httpCall)

	if err != nil {
		return result, callSummary, err
	}

	// now read response into memory, so that we can return the body
	var body []byte
	body, err = ioutil.ReadAll(callSummary.HttpResponse.Body)

	if err != nil {
		return result, callSummary, err
	}

	callSummary.HttpResponseBody = string(body)

	// if result is passed in as nil, it means the API defines no response body
	// json
	if reflect.ValueOf(result).IsValid() && !reflect.ValueOf(result).IsNil() {
		err = json.Unmarshal([]byte(callSummary.HttpResponseBody), &result)
	}

	return result, callSummary, err
}