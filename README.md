# Reddit Bot for Discord

A discord bot that replaces image/gif, text, and video posts with a rich preview  

### [Invite the bot](https://discord.com/oauth2/authorize?client_id=699350209888518244&scope=bot&permissions=59456)

## Image Preview
![image preview](./assets/image.png)

## Video Preview
![image preview](./assets/video.png)

## Text Preview (1000 Character limit)
![image preview](./assets/text.png)

## Build

Requires Go 1.15+

`make all` or `CGO_ENABLED=0 go build -ldflags "-s -w" ./cmd/reddit-bot`