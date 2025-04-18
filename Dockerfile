##########################################
# Builder image to build the application #
##########################################
FROM golang:1.24-alpine AS builder

ARG SRC_PATH

RUN apk add --no-cache upx=4.2.4-r0

WORKDIR /go/src/github.com/${SRC_PATH}

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o builds/app-linux-amd64 . && \
    upx --best --lzma builds/app-linux-amd64

##########################################
# Final image to run the application     #
##########################################
FROM gcr.io/distroless/static-debian12@sha256:3d0f463de06b7ddff27684ec3bfd0b54a425149d0f8685308b1fdf297b0265e9

ARG SRC_PATH

# RUN apk --no-cache add ca-certificates

WORKDIR /app/
COPY --from=builder /go/src/github.com/${SRC_PATH}/builds/app-linux-amd64 ./app

EXPOSE 8080

CMD ["./app"]