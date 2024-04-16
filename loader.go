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
	if argn == 0 {
		return false, -1
	}

	for i := 0; i < argn; i++ {
		if args[i] == "run" && i+1 < argn {
			return true, i + 1
		}
	}
	return false, -1
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
		return
	}

	packStarted := time.Now()

	jsScript, err := k6pack.Pack(string(source), opts)
	if err != nil {
		logrus.WithError(err).Fatal()
		return
	}

	if os.Getenv("XK6_TS_BENCHMARK") == "true" {
		duration := time.Since(packStarted)
		logrus.WithField("extension", "xk6-ts").WithField("duration", duration).Info("Bundling completed in ", duration)
	}

	logrus.WithField("extension", "xk6-ts").Info("Foobar")

	os.Args[scriptIndex] = "-" // Set this so k6 reads from stdin

	reader, writer, err := os.Pipe()
	if err != nil {
		logrus.WithError(err).Fatal()
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Replace os.Stdin with the read end of the pipe
	origStdin := os.Stdin
	os.Stdin = reader


	// Start a goroutine to handle the writing to the pipe
	go func() {
		defer wg.Done()
		defer writer.Close() // Close writer after writing to signal EOF
		if _, err := writer.Write(jsScript); err != nil {
			logrus.WithError(err).Error("Failed to write JS script to pipe")
		}
	}()


	// like a finally
	defer func() {
		os.Stdin = origStdin
		// reader.Close()
	}()

	wg.Wait() // Wait for writing to complete before proceeding

}
