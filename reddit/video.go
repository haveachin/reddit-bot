package reddit

import (
	"os"
	"os/exec"
)

func (p Post) DownloadVideo() (*os.File, error) {
	path := p.ID + ".mp4"
	if err := downloadVideo(path, p.URL()); err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, os.Remove(path)
	}

	return file, nil
}

func downloadVideo(filepath string, url string) error {
	cmd := exec.Command("yt-dlp",
		"--no-continue",
		"--postprocessor-args", "-c:v libx264 -c:a aac -crf 32 -preset faster",
		"-o", filepath,
		url,
	)
	return cmd.Run()
}
