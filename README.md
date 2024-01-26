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

Requires Go 1.21+

`make all` or `CGO_ENABLED=0 go build -ldflags "-s -w" ./cmd/reddit-bot`

## Configuration

See the [default configuration file](configs/config.yml) for more information.

### Post Processing for Videos

The Reddit Bot uses yt-dlp for downloading and processing videos. This allows you to customize the post processing args for your hardware.

#### Intel Hardware Accelleration with VAAPI (i965)

```yml
postProcessingArgs:
  - >-
    Merger+ffmpeg_o:
    -vaapi_device /dev/dri/renderD128
    -vcodec h264_vaapi
    -vf 'format=nv12,hwupload,scale_vaapi=iw/2:ih/2'
    -qp 28
    -fpsmax 24
    -c:a libopus
    -b:a 64k
```
#### Nvidia Hardware Accelleration with CUDA

```yml
postProcessingArgs:
  - >-
    Merger+ffmpeg_i1:
    -hwaccel cuda
    -hwaccel_output_format cuda
  - >-
    Merger+ffmpeg_o:
    -vf 'hwupload,scale_cuda=iw/2:ih/2'
    -c:v h264_nvenc
    -rc constqp
    -qp 28
    -fpsmax 24
    -c:a libopus
    -b:a 64k
```
yt-dlp -o out.mp4 -v --ppa "ffmpeg_i:-hwaccel cuda -hwaccel_output_format cuda -vf 'hwupload,scale_cuda=iw/2:ih/2' -c:v h264_nvenc -rc constqp -qp 28 -fpsmax 24 -c:a libopus -b:a 64k" https://www.reddit.com/1abbkcx

Hey, I'm trying t