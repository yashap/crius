package util

import (
	"bytes"
	"encoding/json"
	"log"
)

func JsonBuffer(rawJson map[string]interface{}) *bytes.Buffer {
	jsonBytes, err := json.Marshal(rawJson)
	if err != nil {
		log.Fatal(err)
	}
	return bytes.NewBuffer(jsonBytes)
}

func JsonString(rawJson map[string]interface{}) string {
	jsonBytes, err := json.Marshal(rawJson)
	if err != nil {
		log.Fatal(err)
	}
	return string(jsonBytes)
}
