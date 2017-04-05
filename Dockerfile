FROM alpine
ADD gopath/bin/gcp-rover /gcp-rover
ENTRYPOINT ["/gcp-rover"]
