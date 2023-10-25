package transformer

import ()

func Suffix(users []string, suffix string) ([]string, error) {
	result := make([]string, len(users))

	for i, user := range users {
		result[i] = user + suffix
	}
	return result, nil

}
