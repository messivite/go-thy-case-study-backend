# Build from repository root:
#   docker build -f Dockerfile .
# Module: github.com/messivite/go-thy-case-study-backend

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/api ./api
ENV PORT=8082
EXPOSE 8082
USER nobody
CMD ["./api"]
