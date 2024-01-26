package reddit

import (
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
)

func (p Post) DownloadVideo() (*os.File, error) {
	path := p.ID + ".mp4"
	if err := p.downloadAndProcessVideo(path, p.ShortURL()); err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, os.Remove(path)
	}

	return file, nil
}

func (p Post) downloadAndProcessVideo(filepath string, url string) error {
	args := []string{
		"--no-continue",
		"-o", filepath,
	}

	if len(p.PostProcessingArgs) != 0 {
		for _, ppas := range p.PostProcessingArgs {
			args = append(args, "--postprocessor-args", ppas)
		}
	}

	cmd := exec.Command("yt-dlp")

	if log.Debug().Enabled() {
		cmd.Stdout = log.Logger
		cmd.Stderr = log.Logger
		args = append(args, "-v")
	}

	args = append(args, url)
	cmd.Args = args

	log.Debug().
		Str("cmd", cmd.String()).
		Msg("yt-dlp cmd")

	return cmd.Run()
}
