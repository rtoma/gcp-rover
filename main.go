package main

import (
	"encoding/json"
	"fmt"
	"html"
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
	w.Write([]byte(fmt.Sprintf("Request from %s\n\n", r.RemoteAddr)))

	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(dump)
}

func curlHandler(w http.ResponseWriter, r *http.Request) {
	url := r.PostFormValue("url")

	if url == "" {
		fmt.Fprint(w, `
<html>
<form method="POST">
URL: <input type="text" name="url" size="80">
<input type="submit" value="Fire!">
</form>
</html>`)
		return
	}
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(w, "GET failed: %s", err)
		return
	}

	fmt.Fprintf(w, "Response to GET %s\n\n", url)
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		fmt.Fprintf(w, "DumpResponse failed: %s", err)
		return
	}
	fmt.Fprint(w, html.EscapeString(string(dump)))
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
	http.HandleFunc("/curl", curlHandler)
	http.ListenAndServe(":8000", nil)
}
