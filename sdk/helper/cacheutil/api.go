package cacheutil

import (
	"net/http"
	"time"
)

type Options struct {
	// Set to True for a private cache, which is not shared among users (eg, in a browser)
	// Set to False for a "shared" cache, which is more common in a server context.
	PrivateCache bool
}

// Given an HTTP Request, the future Status Code, and an ResponseWriter,
// determine the possible reasons a response SHOULD NOT be cached.
func CachableResponseWriter(req *http.Request,
	statusCode int,
	resp http.ResponseWriter,
	opts Options) ([]Reason, time.Time, error) {
	return UsingRequestResponse(req, statusCode, resp.Header(), opts.PrivateCache)
}

// Given an HTTP Request and Response, determine the possible reasons a response SHOULD NOT
// be cached.
func CachableResponse(req *http.Request,
	resp *http.Response,
	opts Options) ([]Reason, time.Time, error) {
	return UsingRequestResponse(req, resp.StatusCode, resp.Header, opts.PrivateCache)
}
