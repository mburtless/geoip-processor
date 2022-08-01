FROM golang:1.18 as builder
WORKDIR /builder
COPY . /builder
RUN CGO_ENABLED=0 go build -o geoip-processor ./cmd/geoip-processor

FROM gcr.io/distroless/static:nonroot
USER nonroot:nonroot
COPY --from=builder /builder/geoip-processor /bin/
ENTRYPOINT ["/bin/geoip-processor"]