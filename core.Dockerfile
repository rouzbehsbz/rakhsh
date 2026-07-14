FROM golang:1.26.3-alpine3.23 AS builder
RUN apk update && apk add --no-cache make git
WORKDIR /app
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.io,direct
ENV GO111MODULE=on CGO_ENABLED=0
RUN go mod download
COPY . .
RUN make build-core

FROM golang:1.26.3-alpine3.23
RUN apk update && apk add --no-cache make git
COPY scripts/install-dep.sh ./scripts/install-dep.sh
RUN chmod +x ./scripts/install-dep.sh
RUN ./scripts/install-dep.sh
RUN adduser -D guard
USER guard
WORKDIR /app
COPY --from=builder /app/.bin/ ./.bin
COPY --from=builder /app/Makefile .
COPY --from=builder /app/db/migrations ./migrations
RUN make migrate-deploy
CMD [".bin/core", "-dev=false"]
