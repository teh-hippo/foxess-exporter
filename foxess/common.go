package foxess

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/teh-hippo/foxess-exporter/util"
)

func IsError(errorNumber int, message string) error {
	if errorNumber != 0 {
		return fmt.Errorf("error response from foxess: %d - %s", errorNumber, message)
	}
	return nil
}

func WriteDebug(message interface{}, name string) {
	out, err := json.MarshalIndent(message, "", "  ")
	if err == nil {
		err = util.ToFile(name, out)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing json: %v\n", err)
	}
}
