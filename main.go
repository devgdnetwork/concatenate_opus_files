package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	"golang.org/x/sync/errgroup"
)

func main() {
	// Grab large PCM file
	largePcmFile, err := os.Open("files/full_raw.pcm")
	if err != nil {
		log.Fatalln(err)
	}

	largePcmFileStat, err := largePcmFile.Stat()
	if err != nil {
		log.Fatalln(err)
	}

	// Split into chunks & encode in parallel
	chunkSize := largePcmFileStat.Size() / int64(runtime.GOMAXPROCS(0))
	chunkSize = int64(roundUpToNearestMultiple(float64(chunkSize), 4))

	g, gCtx := errgroup.WithContext(context.TODO())
	encodedOpusFiles := make([]string, runtime.GOMAXPROCS(0))

	for i := range runtime.GOMAXPROCS(0) {
		chunkSrc := io.NewSectionReader(largePcmFile, int64(i)*chunkSize, chunkSize)

		g.Go(func() error {
			encodedFilePath, err := encodePcmToOpus(gCtx, chunkSrc)
			encodedOpusFiles[i] = encodedFilePath
			return err
		})
	}

	if err := g.Wait(); err != nil {
		log.Fatalln(err)
	}

	// Concatenate chunks
	finalOpusFile, err := concatOpusFiles(context.TODO(), encodedOpusFiles)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("finalOpusFile:", finalOpusFile.Name())
}
