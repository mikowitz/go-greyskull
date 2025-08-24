package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// InputReader provides an abstraction for user input operations,
// making commands testable and flexible in terms of input sources.
type InputReader interface {
	// ReadLine reads a single line of input after displaying the prompt
	ReadLine(prompt string) (string, error)

	// ReadFloat reads and parses a floating-point number after displaying the prompt
	ReadFloat(prompt string) (float64, error)

	// ReadInt reads and parses an integer after displaying the prompt
	ReadInt(prompt string) (int, error)

	// ReadPositiveFloat reads a positive floating-point number, rejecting negative values
	ReadPositiveFloat(prompt string) (float64, error)

	// ReadPositiveInt reads a positive integer, rejecting negative values and zero
	ReadPositiveInt(prompt string) (int, error)
}

// CLIInputReader implements InputReader for command-line interface usage
type CLIInputReader struct {
	in      io.Reader
	out     io.Writer
	scanner *bufio.Scanner
}

// NewCLIInputReader creates a new CLIInputReader with the specified input and output streams
func NewCLIInputReader(in io.Reader, out io.Writer) *CLIInputReader {
	return &CLIInputReader{
		in:      in,
		out:     out,
		scanner: bufio.NewScanner(in),
	}
}

// ReadLine reads a single line of input after displaying the prompt
func (r *CLIInputReader) ReadLine(prompt string) (string, error) {
	// Display the prompt if provided
	if prompt != "" {
		if _, err := r.out.Write([]byte(prompt)); err != nil {
			// Continue even if writing the prompt fails
		}
	}

	// Read a line from input using the persistent scanner
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		return "", fmt.Errorf("no input available")
	}

	// Trim whitespace and return
	return strings.TrimSpace(r.scanner.Text()), nil
}

// ReadFloat reads and parses a floating-point number after displaying the prompt
func (r *CLIInputReader) ReadFloat(prompt string) (float64, error) {
	input, err := r.ReadLine(prompt)
	if err != nil {
		return 0, err
	}

	// Check for empty input
	if input == "" {
		return 0, fmt.Errorf("input cannot be empty")
	}

	// Parse the float
	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %s", input)
	}

	return value, nil
}

// ReadInt reads and parses an integer after displaying the prompt
func (r *CLIInputReader) ReadInt(prompt string) (int, error) {
	input, err := r.ReadLine(prompt)
	if err != nil {
		return 0, err
	}

	// Check for empty input
	if input == "" {
		return 0, fmt.Errorf("input cannot be empty")
	}

	// Parse the integer
	value, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %s", input)
	}

	return value, nil
}

// ReadPositiveFloat reads a positive floating-point number, rejecting negative values
func (r *CLIInputReader) ReadPositiveFloat(prompt string) (float64, error) {
	value, err := r.ReadFloat(prompt)
	if err != nil {
		return 0, err
	}

	// Check if positive
	if value <= 0 {
		return 0, fmt.Errorf("number must be positive, got: %g", value)
	}

	return value, nil
}

// ReadPositiveInt reads a positive integer, rejecting negative values and zero
func (r *CLIInputReader) ReadPositiveInt(prompt string) (int, error) {
	value, err := r.ReadInt(prompt)
	if err != nil {
		return 0, err
	}

	// Check if positive
	if value <= 0 {
		return 0, fmt.Errorf("number must be positive, got: %d", value)
	}

	return value, nil
}

