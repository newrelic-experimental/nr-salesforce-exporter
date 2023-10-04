package deser

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

// JSON deserializer
func DeserJson(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		log.Error("Deser error = ", err.Error())
	}

	return nil
}
