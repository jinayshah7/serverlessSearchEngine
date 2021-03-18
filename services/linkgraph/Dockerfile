########################################################
# STEP 1 use a temporary image to build a static binary
########################################################
FROM golang:1.13 AS builder

# Pull build dependencies
WORKDIR $GOPATH/src/distributedSearchEngine/
COPY . .
#COPY services/linkgraph $GOPATH/src/distributedSearchEngine/services/linkgraph
RUN GO111MODULE=on

RUN go mod download all
RUN go build ./services/linkgraph/main.go

########################################################
# STEP 2 create alpine image with trusted certs
########################################################

#FROM alpine:3.10
#RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
#COPY --from=builder /go/bin/linkgraph /go/bin/linkgraph
#
#ENTRYPOINT ["/go/bin/linkgraph"]
