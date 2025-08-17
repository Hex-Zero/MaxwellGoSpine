# syntax=docker/dockerfile:1
FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /out/app ./cmd/server

FROM gcr.io/distroless/base-debian12
ENV APP_NAME=maxwell-api
COPY --from=build /out/app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
