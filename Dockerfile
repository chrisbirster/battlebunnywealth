FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package*.json ./
RUN npm install
COPY web ./
RUN npm run build

FROM golang:1.27-alpine AS go
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
COPY web ./web
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/battlebunnywealth ./cmd/battlebunnywealth

FROM alpine:3.22
RUN adduser -D -H -u 10001 app
USER app
COPY --from=go /out/battlebunnywealth /usr/local/bin/battlebunnywealth
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/battlebunnywealth"]
