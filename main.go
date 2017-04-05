package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"cloud.google.com/go/compute/metadata"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!")
}

type MetadataResponse struct {
	ProjectID          string
	NumericProjectID   string
	SysHostname        string
	Hostname           string
	InternalIP         string
	ExternalIP         string
	InstanceID         string
	InstanceName       string
	Zone               string
	InstanceAttributes []string
	InstanceTags       []string
	ProjectAttributes  []string
	Scopes             []string
}

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	if !metadata.OnGCE() {
		io.WriteString(w, "not on gce")
		return
	}

	response := &MetadataResponse{}

	v, _ := os.Hostname()
	response.SysHostname = v

	v, _ = metadata.ProjectID()
	response.ProjectID = v

	v, _ = metadata.NumericProjectID()
	response.NumericProjectID = v

	v, _ = metadata.ExternalIP()
	response.ExternalIP = v

	v, _ = metadata.InternalIP()
	response.InternalIP = v

	v, _ = metadata.Hostname()
	response.Hostname = v

	v, _ = metadata.InstanceID()
	response.InstanceID = v

	v, _ = metadata.InstanceName()
	response.InstanceName = v

	v, _ = metadata.Zone()
	response.Zone = v

	vs, _ := metadata.InstanceAttributes()
	response.InstanceAttributes = vs

	vs, _ = metadata.InstanceTags()
	response.InstanceTags = vs

	vs, _ = metadata.ProjectAttributes()
	response.ProjectAttributes = vs

	vs, _ = metadata.Scopes("default")
	response.Scopes = vs

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
