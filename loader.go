package ts

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/szkiba/k6pack"
)

func init() {
	redirectStdin()
}

func isRunCommand(args []string) (bool, int) {
	argn := len(args)

	scriptIndex := argn - 1
	if scriptIndex < 0 {
		return false, scriptIndex
	}

	var runIndex int

	for idx := 0; idx < argn; idx++ {
		arg := args[idx]
		if arg == "run" && runIndex == 0 {
			runIndex = idx

			break
		}
	}

	if runIndex == 0 {
		return false, -1
	}

	return true, scriptIndex
}

func redirectStdin() {
	if os.Getenv("XK6_TS") == "false" {
		return
	}

	isRun, scriptIndex := isRunCommand(os.Args)
	if !isRun {
		return
	}

	filename := os.Args[scriptIndex]
	if filename == "-" {
		return
	}

	opts := &k6pack.Options{
		Filename:  filename,
		SourceMap: os.Getenv("XK6_TS_SOURCEMAP") != "false",
	}

	source, err := os.ReadFile(filepath.Clean(filename))
	if err != nil {
		logrus.WithError(err).Fatal()
	}

	packStarted := time.Now()

	jsScript, err := k6pack.Pack(string(source), opts)
	if err != nil {
		logrus.WithError(err).Fatal()
	}

	if os.Getenv("XK6_TS_BENCHMARK") == "true" {
		duration := time.Since(packStarted)
		logrus.WithField("extension", "xk6-ts").WithField("duration", duration).Info("Bundling completed in ", duration)
	}

	logrus.WithField("extension", "xk6-ts").Info("Creating pipe")
	reader, writer, err := os.Pipe()
	logrus.WithField("extension", "xk6-ts").Info("Created pipe")
	if err != nil {
		logrus.WithError(err).Fatal()
		return
	}

	logrus.WithField("extension", "xk6-ts").Info("Creating wait group")
	var wg sync.WaitGroup
	wg.Add(1)

	// Replace os.Stdin with the read end of the pipe
	logrus.WithField("extension", "xk6-ts").Info("replacing stdin")
	origStdin := os.Stdin
	os.Stdin = reader

	logrus.WithField("extension", "xk6-ts").Info("starting goroutine")
	// Start a goroutine to handle the writing to the pipe
	go func() {
		logrus.WithField("extension", "xk6-ts").Info("inside goroutine")
		defer wg.Done()
		defer writer.Close() // Close writer after writing to signal EOF
		logrus.WithField("extension", "xk6-ts").Info("writer.Write")
		if _, err := writer.Write(jsScript); err != nil {
			logrus.WithField("extension", "xk6-ts").Info("writer.Write errored")
			logrus.WithError(err).Error("Failed to write JS script to pipe")
		}
		logrus.WithField("extension", "xk6-ts").Info("Writer.write completed")
	}()


	defer func() {
		logrus.WithField("extension", "xk6-ts").Info("inside defer/finally")
		os.Stdin = origStdin
		logrus.WithField("extension", "xk6-ts").Info("restored stdin")
		reader.Close()
		logrus.WithField("extension", "xk6-ts").Info("closed reader")
	}()

	logrus.WithField("extension", "xk6-ts").Info("wait")
	wg.Wait() // Wait for writing to complete before proceeding
	logrus.WithField("extension", "xk6-ts").Info("wait done")

	os.Args[scriptIndex] = "-" // Set this so k6 reads from stdin
	logrus.WithField("extension", "xk6-ts").Info("done")
}
