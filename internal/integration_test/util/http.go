package util

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

type HttpResponse struct {
	Code int
	Body map[string]interface{}
}

func HttpRequest(router *gin.Engine, method string, url string, body map[string]interface{}) HttpResponse {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/services", Json(body))

	router.ServeHTTP(w, req)
	jsonMap := make(map[string]interface{})
	bodyString := w.Body.String()
	err := json.Unmarshal([]byte(bodyString), &jsonMap)
	if err != nil {
		return HttpResponse{
			Code: w.Code,
			Body: gin.H{
				"error": "Failed to decode response to JSON",
				"body":  bodyString,
			},
		}
	}
	return HttpResponse{Code: w.Code, Body: jsonMap}
}
