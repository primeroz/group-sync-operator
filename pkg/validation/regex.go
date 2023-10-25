package validation

import (
  "regexp"
  "errors"
  "fmt"
)


func ValidateUsersRegex(users []string, regex string ) error {

  if regex == "" {
    return errors.New("Regular Expression for validation cannot be empty string")
  }

  re, err := regexp.Compile(regex)

  if err != nil {
    return err
  }

  for _, user := range users {
    fmt.Println("DEBUG VALIDATION", user, "REGEX", regex, "RE", re)
    if !re.MatchString(user) {
      return errors.New("Elements of the users list not validating against Validation Regex")
    }
  }

  return nil
}
