package testdata

import (
	"bufio"
	"io"
	"strings"
)

// MockReader simulates user input for testing interactive functions
type MockReader struct {
	inputs []string
	index  int
}

func NewMockReader(inputs ...string) *MockReader {
	return &MockReader{
		inputs: inputs,
		index:  0,
	}
}

func (m *MockReader) ReadString(delim byte) (string, error) {
	if m.index >= len(m.inputs) {
		return "", io.EOF
	}
	input := m.inputs[m.index]
	m.index++
	if !strings.HasSuffix(input, string(delim)) {
		input += string(delim)
	}
	return input, nil
}

// CreateMockReader creates a bufio.Reader from mock inputs
func CreateMockReader(inputs ...string) *bufio.Reader {
	combined := strings.Join(inputs, "\n")
	if !strings.HasSuffix(combined, "\n") {
		combined += "\n"
	}
	return bufio.NewReader(strings.NewReader(combined))
}

// Note: AssertConfigEquals removed due to import cycle
// Use direct field comparison in tests instead
