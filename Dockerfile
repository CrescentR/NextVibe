FROM golang:1.22-alpine AS builder

WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/nextvibe ./cmd/nextvibe

FROM scratch

COPY --from=builder /out/nextvibe /usr/local/bin/nextvibe

ENTRYPOINT ["/usr/local/bin/nextvibe"]
