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
	var req *http.Request
	if len(body) == 0 {
		req, _ = http.NewRequest(method, url, nil)
	} else {
		req, _ = http.NewRequest(method, url, JsonBuffer(body))
	}

	router.ServeHTTP(w, req)
	jsonMap := make(map[string]interface{})
	responseBody := w.Body
	if responseBody == nil {
		return HttpResponse{Code: w.Code, Body: nil}
	}
	bodyString := responseBody.String()
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
