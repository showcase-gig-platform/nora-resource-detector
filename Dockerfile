FROM public.ecr.aws/docker/library/golang:1.19 AS builder

WORKDIR /workdir

COPY . .

RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o nora-resource-detector main.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /workdir/nora-resource-detector .

ENTRYPOINT ["/nora-resource-detector"]
