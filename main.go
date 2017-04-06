package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"cloud.google.com/go/trace"
	"cloud.google.com/go/trace/traceutil"
	"golang.org/x/net/context"
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

var traceClient *trace.Client

func inspectHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("Request from %s\n\n", r.RemoteAddr)))

	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(dump)
}

func envHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Environment:\n\n%s", html.EscapeString(strings.Join(os.Environ(), "\n")))
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

	span := traceClient.NewSpan("curl")
	defer span.Finish()
	span.SetLabel("custom/url", url)

	hc := traceutil.NewHTTPClient(traceClient, nil)
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := hc.Do(req)
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

	log.Print("starting")

	ctx := context.Background()
	if projectId, err := metadata.ProjectID(); err == nil {
		if traceClient, err = trace.NewClient(ctx, projectId); err != nil {
			log.Fatalf("cant create trace client: %s", err)
		}
		log.Printf("enabled tracing, projectid: %s", projectId)
	} else {
		log.Print("cant enable tracing, no metadata available (not in gce?)")
	}

	log.Print("start serving")

	http.Handle("/", traceutil.HTTPHandler(traceClient, helloHandler))
	http.Handle("/metadata", traceutil.HTTPHandler(traceClient, metadataHandler))
	http.Handle("/inspect", traceutil.HTTPHandler(traceClient, inspectHandler))
	http.Handle("/curl", traceutil.HTTPHandler(traceClient, curlHandler))
	http.Handle("/env", traceutil.HTTPHandler(traceClient, envHandler))
	http.ListenAndServe(":8000", nil)
}
