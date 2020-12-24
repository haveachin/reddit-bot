package reddit

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
)

// DownloadVideo downloads the audio and video file from the Reddit servers and
// combines them locally with ffmpeg into one video file.
// This is thread safe and can be executed in parallel.
// Returns the combined video file, event log, and an error
func (video Video) DownloadVideo() (*os.File, []byte, error) {
	eventLog := bytes.NewBuffer([]byte{})
	eventLogger := log.New(eventLog, "", log.Ldate|log.Ltime)

	audioFileName := randomMP4FileName()
	eventLogger.Printf("Downloading audio from \"%s\" into file \"%s\"", video.AudioURL, audioFileName)
	if err := downloadFile(audioFileName, video.AudioURL); err != nil {
		eventLogger.Println(err)
		return nil, eventLog.Bytes(), err
	}
	defer os.Remove(audioFileName)

	videoFileName := randomMP4FileName()
	eventLogger.Printf("Downloading video from \"%s\" into file \"%s\"", video.VideoURL, videoFileName)
	if err := downloadFile(videoFileName, video.VideoURL); err != nil {
		eventLogger.Println(err)
		return nil, eventLog.Bytes(), err
	}
	defer os.Remove(videoFileName)

	outputFileName := randomMP4FileName()
	eventLogger.Printf("Combining audio and video into file \"%s\"", outputFileName)
	if err := combineAudioAndVideo(eventLog, audioFileName, videoFileName, outputFileName); err != nil {
		eventLogger.Println(err)
		os.Remove(outputFileName)
		return nil, eventLog.Bytes(), err
	}

	file, err := os.Open(outputFileName)
	if err != nil {
		eventLogger.Println(err)
		os.Remove(outputFileName)
		return nil, nil, err
	}

	return file, eventLog.Bytes(), nil
}

func randomMP4FileName() string {
	return fmt.Sprintf("%d.mp4", rand.Int())
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		defer os.Remove(filepath)
		return err
	}

	return nil
}

func combineAudioAndVideo(eventLog io.Writer, audioFilePath, videoFilePath, outputFileName string) error {
	cmd := exec.Command("ffmpeg",
		"-i", videoFilePath,
		"-i", audioFilePath,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-crf", "24",
		"-preset", "faster",
		"-tune", "film",
		"-vf", "scale='min(480, iw)':-2",
		outputFileName,
	)
	cmd.Stdout = eventLog
	cmd.Stderr = eventLog
	return cmd.Run()
}
