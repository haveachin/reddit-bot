FROM golang:latest AS builder
LABEL stage=intermediate
COPY . /reddit-bot
WORKDIR /reddit-bot
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o reddit-bot ./cmd/reddit-bot

FROM alpine:latest
RUN apk --no-cache add ca-certificates ffmpeg yt-dlp mesa-dri-gallium mesa-va-gallium intel-media-driver libva-intel-driver linux-firmware-i915
COPY --from=builder /reddit-bot/reddit-bot /usr/bin/reddit-bot
RUN chmod +x /usr/bin/reddit-bot
RUN mkdir /etc/reddit-bot
ENTRYPOINT [ "/usr/bin/reddit-bot", "-c", "/etc/reddit-bot/config.yml" ]
