package policy

import (
	"encoding/json"
	"os"
)

func WriteAlert(file *os.File, alert Alert) error {
	encoder := json.NewEncoder(file)
	
	if err := encoder.Encode(alert); err != nil {
		return err
	}

	return nil
}