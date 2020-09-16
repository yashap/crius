package util

import (
	"bytes"
	"encoding/json"
	"log"
)

func Json(rawJson map[string]interface{}) *bytes.Buffer {
	jsonBytes, err := json.Marshal(rawJson)
	if err != nil {
		log.Fatal(err)
	}
	return bytes.NewBuffer(jsonBytes)
}
