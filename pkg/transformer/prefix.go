package transformer

import ()

func Prefix(users []string, prefix string) ([]string, error) {
	result := make([]string, len(users))

	for i, user := range users {
		result[i] = prefix + user
	}
	return result, nil

}
