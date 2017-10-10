FROM golang:1.9.1

RUN apt-get update && \
  apt-get install -y --no-install-recommends \
    netcat python-pip virtualenv && \
    apt-get clean

WORKDIR /go/src/github.com/bradseefeld/jirabeat/

RUN go get github.com/andygrunwald/go-jira

# No buildable targets in this repo. We just want to download it. The -d flag is supposed
# to allow that, but isnt working. We get an error if there are no buildable targets, so
# we OR with true.
RUN go get github.com/elastic/beats || true

COPY fields.yml jirabeat.reference.yml main.go main_test.go Makefile ./

COPY _meta ./_meta
COPY cmd ./cmd
COPY config ./config
COPY data ./data
COPY beater ./beater

# Run make twice. Once to generate all the configs, and a second time to build the binary.
RUN make && CGO_ENABLED=0 GOOS=linux make

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

RUN mkdir -p /etc/jirabeat

COPY --from=0 /go/src/github.com/bradseefeld/jirabeat/jirabeat.reference.yml /etc/jirabeat/jirabeat.yml
COPY --from=0 /go/src/github.com/bradseefeld/jirabeat/jirabeat .
COPY --from=0 /go/src/github.com/bradseefeld/jirabeat/fields.yml .
COPY --from=0 /go/src/github.com/bradseefeld/jirabeat/_meta _meta

CMD ./jirabeat -c /etc/jirabeat/jirabeat.yml -e
