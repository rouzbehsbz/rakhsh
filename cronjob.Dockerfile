FROM golang:1.26.3-alpine3.23 AS builder
RUN apk update && apk add --no-cache make git
WORKDIR /app
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.io,direct
ENV GO111MODULE=on CGO_ENABLED=0
RUN go mod download
COPY . .
RUN make build-cronjob

FROM golang:1.26.3-alpine3.23
RUN apk update && apk add --no-cache make git
WORKDIR /app
COPY --from=builder /app/bin/ ./bin
COPY --from=builder /app/Makefile .
RUN chmod +x ./bin/cronjob
CMD ["make", "run-cronjob-prod"]