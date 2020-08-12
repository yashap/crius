package main

import (
	"bytes"
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
	// Setup
	a.Initialize(
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"),
	)

	// Run the tests
	testRunExitCode := m.Run()
	cleanDB()

	// Cleanup
	os.Exit(testRunExitCode)
}

func TestEmptyTable(t *testing.T) {
	cleanDB()

	req, _ := http.NewRequest("GET", "/products", nil)
	q := req.URL.Query()
	q.Add("count", "10")
	q.Add("start", "0")
	req.URL.RawQuery = q.Encode()
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	cleanDB()

	req, _ := http.NewRequest("GET", "/product/1", nil)
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

func TestCreateProduct(t *testing.T) {
	cleanDB()

	var jsonStr = []byte(`{"name":"test product", "price":11.22}`)
	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	name := m["name"]
	price := m["price"]
	id := m["id"]

	if name != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got '%v'", name)
	}

	if price != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", price)
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if id != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", id)
	}
}

func TestGetProduct(t *testing.T) {
	cleanDB()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateProduct(t *testing.T) {
	cleanDB()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	jsonStr := []byte(`{"name":"test product - updated name", "price": 11.22}`)
	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	var updatedProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &updatedProduct)

	if updatedProduct["id"] != originalProduct["id"] {
		t.Errorf(
			"Expected id to remain the same (%v). Got %v",
			originalProduct["id"],
			updatedProduct["id"],
		)
	}

	if updatedProduct["name"] == originalProduct["name"] {
		t.Errorf(
			"Expected name to change from %v to %v, but it didn't change. Got %v",
			originalProduct["name"],
			updatedProduct["name"],
			updatedProduct["name"],
		)
	}

	if updatedProduct["price"] == originalProduct["price"] {
		t.Errorf(
			"Expected price to change from %v to %v, but it didn't change. Got %v",
			originalProduct["price"],
			updatedProduct["price"],
			updatedProduct["price"],
		)
	}
}

func TestDeleteProduct(t *testing.T) {
	cleanDB()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNoContent, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
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

func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		_, err := a.DB.Exec(
			`INSERT INTO "products" ("name", "price") VALUES ($1, $2)`,
			fmt.Sprintf("Product %d", i),
			(i+1.0)*10,
		)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func cleanDB() {
	cleanTable("products")
}

func cleanTable(table string) {
	_, err := a.DB.Exec(fmt.Sprintf(`TRUNCATE "%s" RESTART IDENTITY CASCADE`, table))
	if err != nil {
		log.Fatal(err)
	}
}
