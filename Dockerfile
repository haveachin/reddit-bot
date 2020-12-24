FROM golang:latest AS builder
LABEL stage=intermediate
WORKDIR /
COPY . .
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main .

FROM jrottenberg/ffmpeg:4.1-alpine
LABEL maintainer="Hendrik Jonas Schlehlein <hendrik.schlehlein@gmail.com>"
RUN apk --no-cache add ca-certificates
WORKDIR /reddit-bot
COPY --from=builder /main ./
RUN chmod +x ./main
ENTRYPOINT [ "./main" ]