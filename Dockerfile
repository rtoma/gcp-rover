FROM alpine
RUN apk add --update ca-certificates
ADD gopath/bin/gcp-rover /gcp-rover
ENTRYPOINT ["/gcp-rover"]
