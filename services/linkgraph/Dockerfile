FROM golang:1.17 AS builder

WORKDIR .
COPY . .

RUN GO111MODULE=on go mod tidy 
RUN CGO_ENABLED=0 \
    GOARCH=amd64 \
    GOOS=linux \
    go build -ldflags="-w -s" \
    -o /go/bin/linkgraph \
    main.go


FROM gcr.io/distroless/static-debian10@sha256:39256cddc96ed19eefc89c4077138a45855d736029a256787154eafc8e8ebf30
COPY --from=builder /go/bin/linkgraph /

ENTRYPOINT ["/linkgraph"]
