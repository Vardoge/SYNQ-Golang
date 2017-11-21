package test_server

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

func S3Stub() *httptest.Server {
	var resp []byte
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("here in s3 req", r.RequestURI)
		if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			key := r.PostFormValue("key")
			if key != "fakekey" {
				w.Header().Set("Server", "AmazonS3")
				w.Header().Set("X-Amz-Id-2", "vodyoLHQBqirb+3l76iCOoh1Q3Abo8Bm9TntCC1TZso2pL3WGv9aUclvCWloOZynTAEGxNf51hI=")
				w.Header().Set("X-Amz-Request-Id", "9171F45CEDC982B1")
				w.Header().Set("Date", "Fri, 12 May 2017 04:23:53 GMT")
				w.Header().Set("Etag", "9a81d889d4ea7adfa90c9b28b4bbc42f")
				w.Header().Set("Location", key)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		// be default, return error
		resp = LoadSample("aws_err.xml")
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusPreconditionFailed)
		w.Write(resp)
	}))
}
