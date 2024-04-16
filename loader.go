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
func redirectStdinBackup() {
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

	os.Args[scriptIndex] = "-"

	// pair of connected files
	// reader reads from r return bytes written to writer
	reader, writer, err := os.Pipe()
	if err != nil {
		logrus.WithError(err).Fatal()
	}

	// defer closing the writer to the end of the function
	defer writer.Close() //nolint:errcheck


	// save the original stdin
	origStdin := os.Stdin

	// set the reader as the new stdin
	os.Stdin = reader

	// write the jsScript to the writer, which will be read by the reader
	_, err = writer.Write(jsScript)

	if err != nil {
		// if there is an error, close the writer explicitly and reset the stdin
		writer.Close() //nolint:errcheck,gosec

		os.Stdin = origStdin

		logrus.WithError(err).Fatal()
	}
}

// func redirectStdin() {
//     if os.Getenv("XK6_TS") == "false" {
//         return
//     }

//     isRun, scriptIndex := isRunCommand(os.Args)
//     if !isRun {
//         return
//     }

//     filename := os.Args[scriptIndex]
//     if filename == "-" {
//         return
//     }

//     opts := &k6pack.Options{
//         Filename:  filename,
//         SourceMap: os.Getenv("XK6_TS_SOURCEMAP") != "false",
//     }

//     source, err := os.ReadFile(filepath.Clean(filename))
//     if err != nil {
//         logrus.WithError(err).Fatal()
//         return
//     }

//     packStarted := time.Now()

//     jsScript, err := k6pack.Pack(string(source), opts)
//     if err != nil {
//         logrus.WithError(err).Fatal()
//         return
//     }

//     if os.Getenv("XK6_TS_BENCHMARK") == "true" {
//         duration := time.Since(packStarted)
//         logrus.WithField("extension", "xk6-ts").WithField("duration", duration).Info("Bundling completed in ", duration)
//     }

//     os.Args[scriptIndex] = "-" // Set this so k6 reads from stdin

//     reader, writer, err := os.Pipe()
//     if err != nil {
//         logrus.WithError(err).Fatal()
//         return
//     }

//     // Assign the reader to os.Stdin before starting the write goroutine
//     originalStdin := os.Stdin
//     os.Stdin = reader
//     defer func() {
//         os.Stdin = originalStdin
//         reader.Close() // Ensure the reader is closed at the end
//     }()

//     // Start a goroutine to write to the pipe
//     go func() {
//         defer writer.Close() // Ensure the writer is closed after the write
//         if _, err := writer.Write(jsScript); err != nil {
//             logrus.WithError(err).Error("Failed to write JS script to pipe")
//         }
//     }()
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

    jsScript, err := k6pack.Pack(string(source), opts)
    if err != nil {
        logrus.WithError(err).Fatal()
        return
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



// func redirectStdin() {
//     if os.Getenv("XK6_TS") == "false" {
//         return
//     }

//     isRun, scriptIndex := isRunCommand(os.Args)
//     if !isRun {
//         return
//     }

//     filename := os.Args[scriptIndex]
//     if filename == "-" {
//         return
//     }

//     opts := &k6pack.Options{
//         Filename:  filename,
//         SourceMap: os.Getenv("XK6_TS_SOURCEMAP") != "false",
//     }

//     source, err := os.ReadFile(filepath.Clean(filename))
//     if err != nil {
//         logrus.WithError(err).Fatal()
//         return
//     }

//     packStarted := time.Now()

//     jsScript, err := k6pack.Pack(string(source), opts)
//     if err != nil {
//         logrus.WithError(err).Fatal()
//         return
//     }

//     if os.Getenv("XK6_TS_BENCHMARK") == "true" {
//         duration := time.Since(packStarted)
//         logrus.WithField("extension", "xk6-ts").WithField("duration", duration).Info("Bundling completed in ", duration)
//     }

//     os.Args[scriptIndex] = "-" // Set this so k6 reads from stdin

//     reader, writer, err := os.Pipe()
//     if err != nil {
//         logrus.WithError(err).Fatal()
//         return
//     }

//     originalStdin := os.Stdin
//     defer func() {
//         os.Stdin = originalStdin
//         reader.Close()
//         // writer.Close()
//     }()

//     os.Stdin = reader // Redirect stdin for the k6 process

//     // Write to the pipe in a separate goroutine to avoid blocking
// 	// defer writer.Close() // Close the writer to signal EOF to the reader
//     go func() {

//         _, err := writer.Write(jsScript)
//         if err != nil {
//             logrus.WithError(err).Error("Failed to write JS script to pipe")
// 			// writer.Close()
//         }
//     }()
// }


// func redirectStdin() {
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
// 		return
// 	}

// 	packStarted := time.Now()

// 	jsScript, err := k6pack.Pack(string(source), opts)
// 	if err != nil {
// 		logrus.WithError(err).Fatal()
// 		return
// 	}

// 	if os.Getenv("XK6_TS_BENCHMARK") == "true" {
// 		duration := time.Since(packStarted)
// 		logrus.WithField("extension", "xk6-ts").WithField("duration", duration).Info("Bundling completed in ", duration)
// 	}

// 	os.Args[scriptIndex] = "-" // Set this so k6 reads from stdin


// 	// pair of connected files
// 	// reader reads from r return bytes written to writer
// 	reader, writer, err := os.Pipe()
// 	if err != nil {
// 		logrus.WithError(err).Fatal()
// 		return
// 	}
// 	originalStdin := os.Stdin
// 	os.Stdin = reader


// 	// We use a WaitGroup to wait for both goroutines to finish
// 	var wg sync.WaitGroup

// 	// We add 2 to wait for both reading and writing goroutines
// 	wg.Add(2)

// 	defer writer.Close()

// 	// Here we write the JS script to the writer
// 	go func() {
// 		defer wg.Done()
// 		if _, err := writer.Write(jsScript); err != nil {
// 			logrus.WithError(err).Error("Failed to write JS script to pipe")
// 			writer.Close()
// 			os.Stdin = originalStdin
// 		}
// 	}()

// 	// Here we simulate the k6 process by consuming the data from the reader
// 	// and writing it to /dev/null
// 	go func() {
// 		defer wg.Done()
// 		if _, err := io.Copy(os.Stdin, reader); err != nil {
// 			logrus.WithError(err).Error("Failed to read from pipe")
// 		}
// 		reader.Close()
// 	}()


// 	wg.Wait() // Wait for both goroutines to finish

// }