FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /webhook-agent .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=build /webhook-agent /webhook-agent
EXPOSE 3000
ENTRYPOINT ["/webhook-agent"]
