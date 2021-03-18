########################################################
# STEP 1 use a temporary image to build a static binary
########################################################
#FROM golang:1.13 AS builder
#
## Pull build dependencies
#WORKDIR $GOPATH/src/distributedSearchEngine/
#COPY . .
#RUN GO111MODULE=on
#
#RUN go mod download all
#RUN go build ./services/linkgraph/main.go

########################################################
# STEP 2 create alpine image with trusted certs
########################################################

#FROM alpine:3.10
#RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
#COPY --from=builder /go/bin/linkgraph /go/bin/linkgraph
#
#ENTRYPOINT ["/go/bin/linkgraph"]


FROM golang:1.13 AS builder
RUN mkdir $GOPATH/src/distributedSearchEngine
#WORKDIR $GOPATH/src/distributedSearchEngine/
COPY ./../../ $GOPATH/src/distributedSearchEngine
WORKDIR $GOPATH/src/distributedSearchEngine/
RUN GO111MODULE=on
RUN go mod download all
#RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux \
#    go build -ldflags="-w -s" -o /go/bin/linkgraph ./services/linkgraph/main.go

#
#FROM gcr.io/distroless/static-debian10@sha256:39256cddc96ed19eefc89c4077138a45855d736029a256787154eafc8e8ebf30
##RUN apk update && apk add ca-certificates bash && rm -rf /var/cache/apk/*
#COPY --from=builder /go/bin/linkgraph /
#
#ENTRYPOINT ["/go/bin/linkgraph"]
