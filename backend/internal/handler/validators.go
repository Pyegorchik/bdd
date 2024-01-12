package handler

import (
	"errors"
	"fmt"
	"regexp"
)

func addressValidation(address string) error {
	regExp, err := regexp.Compile("^0x[a-fA-F0-9]{40}$")
	if err != nil {
		return fmt.Errorf("addressValidation/regexp.Compile: %w", err)
	}
	if !regExp.MatchString(address) {
		return errors.New("invalid address")
	}
	return nil
}
