# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/maidenmap-server ./cmd/server && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/maidenmap-update-data ./cmd/update-data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/maidenmap-server /usr/local/bin/maidenmap-server
COPY --from=build /out/maidenmap-update-data /usr/local/bin/maidenmap-update-data
WORKDIR /
EXPOSE 8080
USER nonroot:nonroot
ENV DATA_DIR=/data LISTEN_ADDR=:8080
ENTRYPOINT ["/usr/local/bin/maidenmap-server"]
