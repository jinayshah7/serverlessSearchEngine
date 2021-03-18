FROM golang:1.13 AS builder

WORKDIR $GOPATH/src/distributedSearchEngine/
COPY . .

RUN GO111MODULE=on
RUN go mod download all
RUN CGO_ENABLED=0 \
    GOARCH=amd64 \
    GOOS=linux \
    go build -ldflags="-w -s" \
    -o /go/bin/textindexer \
    ./services/textindexer/main.go


FROM gcr.io/distroless/static-debian10@sha256:39256cddc96ed19eefc89c4077138a45855d736029a256787154eafc8e8ebf30
COPY --from=builder /go/bin/textindexer /

ENTRYPOINT ["/textindexer"]
