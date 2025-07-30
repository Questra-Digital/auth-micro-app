package api

import (
	"io"
	"net/http"
)

// ReadResponseBody reads and returns the response body as bytes
func ReadResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
 