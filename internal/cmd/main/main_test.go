package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yashap/crius/internal/app"
)

var a app.App

func TestMain(m *testing.M) {
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"),
	)

	testRunExitCode := m.Run()
	cleanDB()
	os.Exit(testRunExitCode)
}

func TestEmptyTable(t *testing.T) {
	cleanDB()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	cleanDB()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	errorMsg := m["error"]
	if errorMsg != "Product not found" {
		t.Errorf(
			"Expected the 'error' key of the response to be set to 'Product not found'. Got '%s'",
			errorMsg,
		)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func cleanDB() {
	cleanTable("products")
}

func cleanTable(table string) {
	_, err := a.DB.Exec(fmt.Sprintf("TRUNCATE %s RESTART IDENTITY CASCADE", table))
	if err != nil {
		log.Fatal(err)
	}
}
