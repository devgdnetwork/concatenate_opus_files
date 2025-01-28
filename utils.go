package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
)

func roundUpToNearestMultiple(number, multiple float64) float64 {
	if multiple == 0 {
		return number
	}

	remainder := math.Mod(number, multiple)
	if remainder == 0 {
		return number
	}

	return number + multiple - remainder
}

func encodePcmToOpus(ctx context.Context, rawPCM io.Reader) (string, error) {
	f, err := os.CreateTemp("", "*.opus")
	if err != nil {
		return "", fmt.Errorf("os.CreateTemp: %w", err)
	}
	defer f.Close()

	buf := bytes.NewBuffer(nil)
	cmd := exec.CommandContext(ctx,
		"ffmpeg", "-hide_banner", "-y",
		"-ac", "2", "-ar", "48000", "-f", "s16le", "-i", "pipe:",
		"-ac", "2", "-ar", "48000", "-vbr", "off", f.Name())
	cmd.Stdin = rawPCM
	cmd.Stderr = buf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cmd.Run: %w, %s", err, buf.String())
	}

	return f.Name(), nil
}

func concatOpusFiles(ctx context.Context, files []string) (*os.File, error) {
	tmpTxtFile, err := os.CreateTemp("", "*.txt")
	if err != nil {
		return nil, fmt.Errorf("os.CreateTemp: %w", err)
	}
	defer tmpTxtFile.Close()

	for _, file := range files {
		if _, err := tmpTxtFile.WriteString(fmt.Sprintf("file '%s'\n", file)); err != nil {
			return nil, fmt.Errorf("tmpTxtFile.WriteString: %w", err)
		}
	}

	tmpAudioFile, err := os.CreateTemp("", "*.opus")
	if err != nil {
		return nil, fmt.Errorf("os.CreateTemp: %w", err)
	}

	// Step 3: Concat demuxer
	errBuf := bytes.NewBuffer(nil)
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-hide_banner",
		"-f", "concat", "-safe", "0", "-i", tmpTxtFile.Name(),
		"-vbr", "off", "-c", "copy", tmpAudioFile.Name(),
	)
	cmd.Stderr = errBuf
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to concat: %w || %s", err, errBuf.String())
	}

	return tmpAudioFile, nil
}
