FROM public.ecr.aws/docker/library/golang:1.19 AS builder

WORKDIR /workdir

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o nora-resource-detector main.go

FROM scratch

WORKDIR /
COPY --from=builder /workdir/nora-resource-detector .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/nora-resource-detector"]
