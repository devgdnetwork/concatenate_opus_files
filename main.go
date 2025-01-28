package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	// Grab large PCM file
	largePcmFile, err := os.Open("files/full_raw.pcm")
	if err != nil {
		log.Fatalln(err)
	}

	// Split into 2 chunks
	ByteRate := 2
	SampleRate := 48000
	Channels := 2
	Seconds := 20
	chunkSize := Seconds * Channels * SampleRate * ByteRate

	file1, err := encodePcmToOpus(context.TODO(), io.LimitReader(largePcmFile, int64(chunkSize)))
	if err != nil {
		log.Fatalln(err)
	}

	file2, err := encodePcmToOpus(context.TODO(), io.LimitReader(largePcmFile, int64(chunkSize)))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Check if these play with no defects:", file1)
	fmt.Println("file1:", file1)
	fmt.Println("file2:", file2)
	fmt.Println()

	concatFile, err := concatOpusFiles(context.TODO(), []string{file1, file2})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("concatted file:", concatFile.Name())
}
