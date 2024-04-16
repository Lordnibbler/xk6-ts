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
// func redirectStdinBackup() {
// 	if os.Getenv("XK6_TS") == "false" {
// 		return
// 	}

// 	isRun, scriptIndex := isRunCommand(os.Args)
// 	if !isRun {
// 		return
// 	}

// 	filename := os.Args[scriptIndex]
// 	if filename == "-" {
// 		return
// 	}

// 	opts := &k6pack.Options{
// 		Filename:  filename,
// 		SourceMap: os.Getenv("XK6_TS_SOURCEMAP") != "false",
// 	}

// 	source, err := os.ReadFile(filepath.Clean(filename))
// 	if err != nil {
// 		logrus.WithError(err).Fatal()
// 	}

// 	packStarted := time.Now()

// 	jsScript, err := k6pack.Pack(string(source), opts)
// 	if err != nil {
// 		logrus.WithError(err).Fatal()
// 	}

// 	if os.Getenv("XK6_TS_BENCHMARK") == "true" {
// 		duration := time.Since(packStarted)
// 		logrus.WithField("extension", "xk6-ts").WithField("duration", duration).Info("Bundling completed in ", duration)
// 	}

// 	os.Args[scriptIndex] = "-"

// 	// pair of connected files
// 	// reader reads from r return bytes written to writer
// 	reader, writer, err := os.Pipe()
// 	if err != nil {
// 		logrus.WithError(err).Fatal()
// 	}

// 	// defer closing the writer to the end of the function
// 	defer writer.Close() //nolint:errcheck


// 	// save the original stdin
// 	origStdin := os.Stdin

// 	// set the reader as the new stdin
// 	os.Stdin = reader

// 	// write the jsScript to the writer, which will be read by the reader
// 	_, err = writer.Write(jsScript)

// 	if err != nil {
// 		// if there is an error, close the writer explicitly and reset the stdin
// 		writer.Close() //nolint:errcheck,gosec

// 		os.Stdin = origStdin

// 		logrus.WithError(err).Fatal()
// 	}
// }

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

    reader, writer, err := os.Pipe()
    if err != nil {
        logrus.WithError(err).Fatal()
        return
    }

    // Set up os.Stdin to be the reader before any writing happens
    originalStdin := os.Stdin
    os.Stdin = reader
    defer func() {
        os.Stdin = originalStdin
        writer.Close() // Ensure the writer is closed when done
        reader.Close() // Ensure the reader is also closed
    }()

    // Write to the pipe in a goroutine
    go func() {
        defer writer.Close() // Close writer after writing to signal EOF
        _, writeErr := writer.Write(jsScript)
        if writeErr != nil {
            logrus.WithError(writeErr).Error("Failed to write JS script to pipe")
        }
    }()
}
