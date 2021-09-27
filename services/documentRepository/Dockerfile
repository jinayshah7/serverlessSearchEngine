FROM golang:1.17 AS builder

WORKDIR $GOPATH/src/distributedSearchEngine/
COPY . .

RUN go mod tidy 
RUN CGO_ENABLED=0 \
    GOARCH=amd64 \
    GOOS=linux \
    go build -ldflags="-w -s" \
    -o /go/bin/documentRepository \
    .


FROM golang:1.17
COPY --from=builder /go/bin/documentRepository /

ENTRYPOINT ["/documentRepository"]
