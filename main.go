package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
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
	InstanceAttributes map[string]string
	InstanceTags       []string
	ProjectAttributes  map[string]string
	Scopes             []string
}

func inspectHandler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(dump)

}

func metadataHandler(w http.ResponseWriter, r *http.Request) {
	if !metadata.OnGCE() {
		io.WriteString(w, "not on gce")
		return
	}

	response := &MetadataResponse{
		InstanceAttributes: make(map[string]string),
		ProjectAttributes:  make(map[string]string),
	}

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
	for _, attr := range vs {
		if val, err := metadata.InstanceAttributeValue(attr); err == nil {
			response.InstanceAttributes[attr] = val
		}
	}

	vs, _ = metadata.InstanceTags()
	response.InstanceTags = vs

	vs, _ = metadata.ProjectAttributes()
	for _, attr := range vs {
		if val, err := metadata.ProjectAttributeValue(attr); err == nil {
			response.ProjectAttributes[attr] = val
		}
	}

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
	http.HandleFunc("/inspect", inspectHandler)
	http.ListenAndServe(":8000", nil)
}
