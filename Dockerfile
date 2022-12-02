FROM golang:1.19.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /go/bin/app

FROM busybox

COPY --from=builder /go/bin/app /go/bin/app

ENV PORT 8080
ENV DB_PATH /data/gol.db
RUN mkdir -p /data

ENTRYPOINT ["/go/bin/app"]
