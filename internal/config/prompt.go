package config

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func ConfirmInsecure(in io.Reader, out io.Writer) (bool, error) {
	fmt.Fprint(out, "No API key provided. Run the proxy without security? [y/N]: ")
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
