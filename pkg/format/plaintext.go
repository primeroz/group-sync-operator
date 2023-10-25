package format

import (
	"bufio"
	"bytes"
)

func ParseUsersFromPlaintext(body []byte) ([]string, error) {
	// Assuming each line in the response body is a user
	users := []string{}
	scanner := bufio.NewScanner(bytes.NewReader(body))
	for scanner.Scan() {
		line := scanner.Text()
		users = append(users, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
