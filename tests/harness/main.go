package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type harnessOptions struct {
	gogoGbPath string
	romPath    string
	success    string
	failure    string
	timeout    time.Duration
	model      string
}

func parseOpts(args []string) (*harnessOptions, error) {
	opts := &harnessOptions{}

	fs := flag.NewFlagSet("gb-test-harness", flag.ContinueOnError)
	fs.StringVar(&opts.success, "success", "", "Substring in the serial output indicating success (required)")
	fs.StringVar(&opts.failure, "failure", "", "Substring in the serial output indicating failure (required)")
	fs.DurationVar(&opts.timeout, "timeout", 30*time.Second, "Timeout for success/failure output")
	fs.StringVar(&opts.model, "model", "", "Console model to pass through to gogo-gb")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage:\n  %s GOGO-GB ROM_FILE\nFlags:\n", os.Args[0])
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parsing args: %w", err)
	}

	if fs.NArg() < 2 {
		return nil, errors.New("expected 2 args")
	}

	gogoGbPath := fs.Arg(0)
	rom := fs.Arg(1)

	if gogoGbPath == "" {
		return nil, errors.New("path to gogo-gb required as first arg")
	}
	opts.gogoGbPath = gogoGbPath

	if rom == "" {
		return nil, errors.New("path to ROM required as second arg")
	}
	opts.romPath = rom

	if opts.success == "" {
		return nil, errors.New("--success is required")
	}

	if opts.failure == "" {
		return nil, errors.New("--failure is required")
	}

	return opts, nil
}

const (
	exitSuccess int = iota
	exitFailure
	exitProcessExited
	exitError
)

type serialResult struct {
	exitCode int
	err      error
}

func pollSerial(
	r io.Reader,
	serialBuf *bytes.Buffer,
	success string,
	failure string,
) serialResult {
	buf := make([]byte, 256)

	for {
		n, err := r.Read(buf)
		if n > 0 {
			serialBuf.Write(buf[:n])

			if bytes.Contains(serialBuf.Bytes(), []byte(failure)) {
				return serialResult{exitCode: exitFailure}
			}
			if bytes.Contains(serialBuf.Bytes(), []byte(success)) {
				return serialResult{exitCode: exitSuccess}
			}
		}

		if err != nil {
			if err == io.EOF {
				return serialResult{exitCode: exitProcessExited}
			}

			return serialResult{exitCode: exitProcessExited, err: err}
		}
	}
}

func logOutputBuffers(w io.Writer, serialBuf *bytes.Buffer, logBuf *bytes.Buffer) {
	fmt.Fprintf(w, "--- serial output ---\n%s\n", serialBuf.String())
	fmt.Fprintf(w, "--- gogo-gb log (stderr) ---\n%s\n", logBuf.String())
}

func run(opts *harnessOptions, stdout, stderr io.Writer) int {
	ctx, cancel := context.WithTimeout(context.Background(), opts.timeout)
	defer cancel()

	args := []string{
		"run",
		opts.romPath,
		"--headless",
		"--serial-port=stdout",
		"--log=stderr",
	}

	if opts.model != "" {
		args = append(args, "--model="+opts.model)
	}

	cmd := exec.CommandContext(ctx, opts.gogoGbPath, args...)

	serialPipe, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: unable to attach to gogo-gb stdout: %v\n", err)

		return exitError
	}

	var logBuf bytes.Buffer
	cmd.Stderr = &logBuf

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(stderr, "ERROR: unable to start gogo-gb: %v\n", err)

		return exitError
	}

	var serialBuf bytes.Buffer
	serialChan := make(chan serialResult, 1)
	go func() {
		serialChan <- pollSerial(serialPipe, &serialBuf, opts.success, opts.failure)
	}()

	var result serialResult
	select {
	case result = <-serialChan:
		err := cmd.Cancel()
		if err != nil {
			fmt.Fprintf(stderr, "WARN: while canceling child: %v\n", err)
		}
	case <-ctx.Done():
		result.exitCode = exitError
		result.err = ctx.Err()
	}

	cmd.Wait()

	switch result.exitCode {
	case exitSuccess:
		fmt.Fprintf(stdout, "PASS: %s\n", opts.romPath)

		return exitSuccess
	case exitFailure:
		fmt.Fprintf(stdout, "FAIL: %s\n", opts.romPath)
		logOutputBuffers(stdout, &serialBuf, &logBuf)

		return exitFailure
	default:
		fmt.Fprintf(stdout, "ERROR: %s exited without a result: %v\n", opts.romPath, result.err)
		logOutputBuffers(stdout, &serialBuf, &logBuf)

		return exitError
	}
}

func main() {
	opts, err := parseOpts(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(exitError)
	}

	exitCode := run(opts, os.Stdout, os.Stderr)

	os.Exit(exitCode)
}
