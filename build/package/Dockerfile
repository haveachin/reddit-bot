FROM golang:latest AS builder
LABEL stage=intermediate
COPY . /reddit-bot
WORKDIR /reddit-bot
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main .

FROM jrottenberg/ffmpeg:4.1-alpine
LABEL maintainer="haveachin <haveachin@proton.me>"
LABEL org.opencontainers.image.source https://github.com/haveachin/reddit-bot
RUN apk --no-cache add ca-certificates
WORKDIR /reddit-bot
COPY --from=builder /main ./
RUN chmod +x ./main
ENTRYPOINT [ "./main" ]