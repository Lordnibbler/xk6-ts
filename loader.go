// Package ts contains xk6-ts extension.
package ts

import (
	"os"
	"path/filepath"
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

//nolint:forbidigo
func redirectStdin() {
	logrus.WithField("extension", "xk6-ts").Info("redirectStdin")

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

	logrus.WithField("extension", "xk6-ts").Info("Bundling completed", string(jsScript), len(jsScript))

	if os.Getenv("XK6_TS_BENCHMARK") == "true" {
		duration := time.Since(packStarted)
		logrus.WithField("extension", "xk6-ts").WithField("duration", duration).Info("Bundling completed in ", duration)
	}

	os.Args[scriptIndex] = "-"

	_, writer, err := os.Pipe()
	if err != nil {
		logrus.WithError(err).Fatal()
	}

	// var wg sync.WaitGroup
	// wg.Add(1)

	// thread/async fn
	// go func() {
	// 	defer func() {
	// 		closeErr := reader.Close()
	// 		if closeErr != nil {
	// 			logrus.WithError(closeErr).Error("Failed to close reader")
	// 		}
	// 		wg.Done()
	// 	}()

	// 	// Read to EOF to ensure all data is consumed.
	// 	_, copyErr := io.Copy(os.Stdout, reader)
	// 	if copyErr != nil {
	// 		logrus.WithError(copyErr).Error("Failed to read from pipe")
	// 	}
	// 	logrus.Info("Reading from pipe completed")
	// }()

	// defer func() {

	// 	// wg.Wait() // Wait for the reading goroutine to finish
	// }()

	logrus.WithField("extension", "xk6-ts").Info("Writing to writer")
	var bytesWritten int
	if bytesWritten, err = writer.Write(jsScript); err != nil {
		logrus.WithField("extension", "xk6-ts").WithError(err).Fatal("Failed to write JS script to pipe")
	}
	logrus.WithField("extension", "xk6-ts").Info("Write completed", bytesWritten)
	closeErr := writer.Close()
	logrus.WithField("extension", "xk6-ts").Info("Writer closed")
	if closeErr != nil {
		logrus.WithField("extension", "xk6-ts").WithError(closeErr).Fatal("Failed to close writer")
	}
}
