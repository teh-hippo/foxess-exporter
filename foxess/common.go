package foxess

import (
	"fmt"
)

func IsError(errorNumber int, message string) error {
	if errorNumber != 0 {
		return fmt.Errorf("error response from foxess: %d - %s", errorNumber, message)
	}
	return nil
}
