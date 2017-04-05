package main

import (
	"encoding/json"
	"io"
	"net/http"

	"cloud.google.com/go/compute/metadata"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!")
}

type MetadataResponse struct {
	ProjectID        string
	NumericProjectID string
}

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	if !metadata.OnGCE() {
		io.WriteString(w, "not on gce")
		return
	}

	response := &MetadataResponse{}
	v, _ := metadata.ProjectID()
	response.ProjectID = v
	v, _ = metadata.NumericProjectID()
	response.NumericProjectID = v

	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/metadata", metadataHandler)
	http.ListenAndServe(":8000", nil)
}
