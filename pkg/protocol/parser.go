package protocol

import (
	"fmt"
	"strings"
)

func ParseCommand(input string) (string, []string, error) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil, fmt.Errorf("empty command")
	}
	return parts[0], parts[1:], nil
}
