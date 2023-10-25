package transformer

import (
	"errors"
	"regexp"
)

func RegexRemove(users []string, regex string) ([]string, error) {

	if regex == "" {
		return nil, errors.New("Regular Expression for RegexRemove transfomer cannot be empty string")
	}

	re, err := regexp.Compile(regex)

	if err != nil {
		return nil, err
	}

	var result []string

	for _, user := range users {
		if !re.MatchString(user) {
			result = append(result, user)
		}
	}

	return result, nil

}
