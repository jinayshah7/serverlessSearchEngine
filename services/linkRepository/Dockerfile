FROM golang:1.17 AS builder

WORKDIR /app 
COPY . .
RUN go mod tidy 
RUN GOARCH=amd64 \
    GOOS=linux \
    go build -ldflags="-w -s" \
    -o /go/bin/linkRepository \
    . 


FROM golang:1.17
COPY --from=builder /go/bin/linkRepository /

ENTRYPOINT ["/linkRepository"]
