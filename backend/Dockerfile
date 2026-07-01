FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /tg-drive .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /tg-drive /tg-drive
EXPOSE 8080
ENTRYPOINT ["/tg-drive"]
