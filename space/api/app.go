package api

import "net/http"

//RedirectGet RedirectGet
func RedirectGet(path string, query []byte) *http.Response {

	resp, err := http.Get(path)
	if err != nil {

	}
	return resp
}
