FROM golang:1.16-alpine as builder
WORKDIR /app
COPY . .

# Compile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o k8s-secret-sync

FROM busybox:latest
COPY --from=builder /app/k8s-secret-sync /k8s-secret-sync
CMD ["/k8s-secret-sync", "/opt/k8s-secret-sync/config.yaml"]
