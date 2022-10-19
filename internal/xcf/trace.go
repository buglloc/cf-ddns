package xcf

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
)

func IPFromTrace(trace []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(trace))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "ip=") {
			continue
		}

		return line[3:], nil
	}

	return "", errors.New("no ip= found trace the trace")
}
