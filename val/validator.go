package val

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	isValidUserName = regexp.MustCompile(`^[a-z0-9_]+$`).MatchString
	// \\s is to space char
	isValidFullName = regexp.MustCompile(`^[a-zA-Z\\s]+$`).MatchString
)

func ValidateString(value string, minLenght, maxLenght int) error {
	n := len(value)
	if n < minLenght || n > maxLenght {
		return fmt.Errorf("must contain from %d-%d characters", minLenght, maxLenght)
	}
	return nil
}

func ValidateUsername(username string) error {
	if err := ValidateString(username, 3, 100); err != nil {
		return err
	}

	if !isValidUserName(username) {
		return fmt.Errorf("must contain only lowercase letters, digits or underscore")
	}
	return nil
}

func ValidatePassword(pass string) error {
	if err := ValidateString(pass, 6, 100); err != nil {
		return err
	}
	return nil
}

func ValidateEmail(in string) error {
	if err := ValidateString(in, 3, 100); err != nil {
		return err
	}

	if _, err := mail.ParseAddress(in); err != nil {
		return fmt.Errorf("is not valid email address")
	}
	return nil
}

func ValidateFullName(name string) error {
	if err := ValidateString(name, 3, 100); err != nil {
		return err
	}
	if !isValidFullName(name) {
		return fmt.Errorf("must contain only letters or spaces")
	}
	return nil
}
