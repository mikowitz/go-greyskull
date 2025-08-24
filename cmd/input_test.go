package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCLIInputReader_ReadLine tests the basic line reading functionality
func TestCLIInputReader_ReadLine(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		prompt         string
		expectedOutput string
		expectedResult string
		shouldError    bool
	}{
		{
			name:           "simple line input",
			input:          "hello world\n",
			prompt:         "Enter text: ",
			expectedOutput: "Enter text: ",
			expectedResult: "hello world",
			shouldError:    false,
		},
		{
			name:           "empty line input",
			input:          "\n",
			prompt:         "Enter something: ",
			expectedOutput: "Enter something: ",
			expectedResult: "",
			shouldError:    false,
		},
		{
			name:           "line with whitespace",
			input:          "  spaced input  \n",
			prompt:         "Input: ",
			expectedOutput: "Input: ",
			expectedResult: "spaced input", // Should be trimmed
			shouldError:    false,
		},
		{
			name:           "multiline input - only first line",
			input:          "first line\nsecond line\n",
			prompt:         "First: ",
			expectedOutput: "First: ",
			expectedResult: "first line",
			shouldError:    false,
		},
		{
			name:        "EOF error",
			input:       "", // No input
			prompt:      "Enter: ",
			shouldError: true,
		},
		{
			name:           "no prompt",
			input:          "input\n",
			prompt:         "",
			expectedOutput: "",
			expectedResult: "input",
			shouldError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			input := strings.NewReader(tt.input)
			
			reader := NewCLIInputReader(input, &output)
			
			result, err := reader.ReadLine(tt.prompt)
			
			if tt.shouldError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedOutput, output.String())
		})
	}
}

// TestCLIInputReader_ReadFloat tests floating-point number parsing
func TestCLIInputReader_ReadFloat(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		prompt         string
		expectedOutput string
		expectedResult float64
		shouldError    bool
		errorContains  string
	}{
		{
			name:           "valid integer",
			input:          "42\n",
			prompt:         "Enter number: ",
			expectedOutput: "Enter number: ",
			expectedResult: 42.0,
			shouldError:    false,
		},
		{
			name:           "valid float",
			input:          "3.14159\n",
			prompt:         "Enter pi: ",
			expectedOutput: "Enter pi: ",
			expectedResult: 3.14159,
			shouldError:    false,
		},
		{
			name:           "negative float",
			input:          "-2.5\n",
			prompt:         "Number: ",
			expectedOutput: "Number: ",
			expectedResult: -2.5,
			shouldError:    false,
		},
		{
			name:           "zero",
			input:          "0\n",
			prompt:         "Zero: ",
			expectedOutput: "Zero: ",
			expectedResult: 0.0,
			shouldError:    false,
		},
		{
			name:          "invalid text",
			input:         "not a number\n",
			prompt:        "Number: ",
			shouldError:   true,
			errorContains: "invalid number",
		},
		{
			name:          "empty input",
			input:         "\n",
			prompt:        "Number: ",
			shouldError:   true,
			errorContains: "input cannot be empty",
		},
		{
			name:          "whitespace only",
			input:         "   \n",
			prompt:        "Number: ",
			shouldError:   true,
			errorContains: "input cannot be empty",
		},
		{
			name:          "mixed valid/invalid",
			input:         "12.5abc\n",
			prompt:        "Number: ",
			shouldError:   true,
			errorContains: "invalid number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			input := strings.NewReader(tt.input)
			
			reader := NewCLIInputReader(input, &output)
			
			result, err := reader.ReadFloat(tt.prompt)
			
			if tt.shouldError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedOutput, output.String())
		})
	}
}

// TestCLIInputReader_ReadInt tests integer parsing
func TestCLIInputReader_ReadInt(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		prompt         string
		expectedOutput string
		expectedResult int
		shouldError    bool
		errorContains  string
	}{
		{
			name:           "positive integer",
			input:          "42\n",
			prompt:         "Enter number: ",
			expectedOutput: "Enter number: ",
			expectedResult: 42,
			shouldError:    false,
		},
		{
			name:           "negative integer",
			input:          "-17\n",
			prompt:         "Negative: ",
			expectedOutput: "Negative: ",
			expectedResult: -17,
			shouldError:    false,
		},
		{
			name:           "zero",
			input:          "0\n",
			prompt:         "Zero: ",
			expectedOutput: "Zero: ",
			expectedResult: 0,
			shouldError:    false,
		},
		{
			name:           "large integer",
			input:          "999999\n",
			prompt:         "Big: ",
			expectedOutput: "Big: ",
			expectedResult: 999999,
			shouldError:    false,
		},
		{
			name:          "float input",
			input:         "3.14\n",
			prompt:        "Integer: ",
			shouldError:   true,
			errorContains: "invalid integer",
		},
		{
			name:          "text input",
			input:         "not a number\n",
			prompt:        "Integer: ",
			shouldError:   true,
			errorContains: "invalid integer",
		},
		{
			name:          "empty input",
			input:         "\n",
			prompt:        "Integer: ",
			shouldError:   true,
			errorContains: "input cannot be empty",
		},
		{
			name:          "mixed number and text",
			input:         "42abc\n",
			prompt:        "Integer: ",
			shouldError:   true,
			errorContains: "invalid integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			input := strings.NewReader(tt.input)
			
			reader := NewCLIInputReader(input, &output)
			
			result, err := reader.ReadInt(tt.prompt)
			
			if tt.shouldError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedOutput, output.String())
		})
	}
}

// TestCLIInputReader_ReadPositiveFloat tests positive float validation
func TestCLIInputReader_ReadPositiveFloat(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		prompt         string
		expectedOutput string
		expectedResult float64
		shouldError    bool
		errorContains  string
	}{
		{
			name:           "positive float",
			input:          "3.14\n",
			prompt:         "Weight: ",
			expectedOutput: "Weight: ",
			expectedResult: 3.14,
			shouldError:    false,
		},
		{
			name:           "positive integer",
			input:          "25\n",
			prompt:         "Reps: ",
			expectedOutput: "Reps: ",
			expectedResult: 25.0,
			shouldError:    false,
		},
		{
			name:           "small positive number",
			input:          "0.1\n",
			prompt:         "Small: ",
			expectedOutput: "Small: ",
			expectedResult: 0.1,
			shouldError:    false,
		},
		{
			name:          "zero",
			input:         "0\n",
			prompt:        "Positive: ",
			shouldError:   true,
			errorContains: "must be positive",
		},
		{
			name:          "negative number",
			input:         "-5.5\n",
			prompt:        "Positive: ",
			shouldError:   true,
			errorContains: "must be positive",
		},
		{
			name:          "invalid text",
			input:         "not a number\n",
			prompt:        "Positive: ",
			shouldError:   true,
			errorContains: "invalid number",
		},
		{
			name:          "empty input",
			input:         "\n",
			prompt:        "Positive: ",
			shouldError:   true,
			errorContains: "input cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			input := strings.NewReader(tt.input)
			
			reader := NewCLIInputReader(input, &output)
			
			result, err := reader.ReadPositiveFloat(tt.prompt)
			
			if tt.shouldError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedOutput, output.String())
		})
	}
}

// TestCLIInputReader_ReadPositiveInt tests positive integer validation
func TestCLIInputReader_ReadPositiveInt(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		prompt         string
		expectedOutput string
		expectedResult int
		shouldError    bool
		errorContains  string
	}{
		{
			name:           "positive integer",
			input:          "42\n",
			prompt:         "Count: ",
			expectedOutput: "Count: ",
			expectedResult: 42,
			shouldError:    false,
		},
		{
			name:           "small positive integer",
			input:          "1\n",
			prompt:         "One: ",
			expectedOutput: "One: ",
			expectedResult: 1,
			shouldError:    false,
		},
		{
			name:           "large positive integer",
			input:          "999\n",
			prompt:         "Big: ",
			expectedOutput: "Big: ",
			expectedResult: 999,
			shouldError:    false,
		},
		{
			name:          "zero",
			input:         "0\n",
			prompt:        "Positive: ",
			shouldError:   true,
			errorContains: "must be positive",
		},
		{
			name:          "negative integer",
			input:         "-5\n",
			prompt:        "Positive: ",
			shouldError:   true,
			errorContains: "must be positive",
		},
		{
			name:          "float input",
			input:         "3.14\n",
			prompt:        "Integer: ",
			shouldError:   true,
			errorContains: "invalid integer",
		},
		{
			name:          "text input",
			input:         "not a number\n",
			prompt:        "Integer: ",
			shouldError:   true,
			errorContains: "invalid integer",
		},
		{
			name:          "empty input",
			input:         "\n",
			prompt:        "Integer: ",
			shouldError:   true,
			errorContains: "input cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			input := strings.NewReader(tt.input)
			
			reader := NewCLIInputReader(input, &output)
			
			result, err := reader.ReadPositiveInt(tt.prompt)
			
			if tt.shouldError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedOutput, output.String())
		})
	}
}

// TestCLIInputReader_ErrorReader tests behavior with failing readers
func TestCLIInputReader_ErrorReader(t *testing.T) {
	// Create a reader that always returns an error
	errorReader := &erroringReader{err: errors.New("read error")}
	
	var output bytes.Buffer
	reader := NewCLIInputReader(errorReader, &output)
	
	// Test all methods handle read errors properly
	_, err := reader.ReadLine("prompt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read error")
	
	_, err = reader.ReadFloat("prompt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read error")
	
	_, err = reader.ReadInt("prompt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read error")
	
	_, err = reader.ReadPositiveFloat("prompt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read error")
	
	_, err = reader.ReadPositiveInt("prompt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read error")
}

// TestCLIInputReader_Constructor tests the constructor
func TestCLIInputReader_Constructor(t *testing.T) {
	var input bytes.Buffer
	var output bytes.Buffer
	
	reader := NewCLIInputReader(&input, &output)
	
	assert.NotNil(t, reader)
	assert.Equal(t, &input, reader.in)
	assert.Equal(t, &output, reader.out)
	assert.NotNil(t, reader.scanner)
}

// TestCLIInputReader_InterfaceCompliance verifies CLIInputReader implements InputReader
func TestCLIInputReader_InterfaceCompliance(t *testing.T) {
	var input bytes.Buffer
	var output bytes.Buffer
	
	// This should compile - verifies interface compliance
	var _ InputReader = NewCLIInputReader(&input, &output)
}

// TestCLIInputReader_MultipleReads tests that the reader can be used multiple times
func TestCLIInputReader_MultipleReads(t *testing.T) {
	input := strings.NewReader("first line\nsecond line\n42\n3.14\n5\n")
	var output bytes.Buffer
	
	reader := NewCLIInputReader(input, &output)
	
	// Read multiple values in sequence
	line1, err := reader.ReadLine("First: ")
	require.NoError(t, err)
	assert.Equal(t, "first line", line1)
	
	line2, err := reader.ReadLine("Second: ")
	require.NoError(t, err)
	assert.Equal(t, "second line", line2)
	
	int1, err := reader.ReadInt("Int: ")
	require.NoError(t, err)
	assert.Equal(t, 42, int1)
	
	float1, err := reader.ReadFloat("Float: ")
	require.NoError(t, err)
	assert.Equal(t, 3.14, float1)
	
	pos1, err := reader.ReadPositiveInt("Positive: ")
	require.NoError(t, err)
	assert.Equal(t, 5, pos1)
	
	// Verify all prompts were written
	expectedOutput := "First: Second: Int: Float: Positive: "
	assert.Equal(t, expectedOutput, output.String())
}

// TestCLIInputReader_WriteError tests behavior when output writing fails
func TestCLIInputReader_WriteError(t *testing.T) {
	input := strings.NewReader("test\n")
	errorWriter := &erroringWriter{err: errors.New("write error")}
	
	reader := NewCLIInputReader(input, errorWriter)
	
	// Should still work even if prompt writing fails
	result, err := reader.ReadLine("This will fail to write: ")
	require.NoError(t, err) // Reading should still succeed
	assert.Equal(t, "test", result)
}

// Helper types for testing error conditions

type erroringReader struct {
	err error
}

func (e *erroringReader) Read(p []byte) (int, error) {
	return 0, e.err
}

type erroringWriter struct {
	err error
}

func (e *erroringWriter) Write(p []byte) (int, error) {
	return 0, e.err
}