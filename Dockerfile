FROM oven/bun:1-alpine AS frontend-builder

WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run build


FROM golang:1.23-alpine AS go-builder

WORKDIR /backend
COPY backend/go.mod backend/go.sum* ./
RUN go mod download

COPY backend/ .
COPY --from=frontend-builder /app/build ./static/files/

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /hopper .


FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && mkdir -p /data

COPY --from=go-builder /hopper /usr/local/bin/hopper

EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["hopper"]
