FROM golang:alpine AS build

RUN apk --no-cache add ca-certificates

WORKDIR /go/src/github.com/JawadMM/ecomm

COPY go.mod go.sum ./
RUN go mod download

COPY events events
COPY catalog catalog

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/app ./catalog/cmd/catalog

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /go/bin/app /usr/bin/app
EXPOSE 8080
CMD ["app"]
