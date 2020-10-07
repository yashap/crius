package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Error struct {
	Message    string
	Details    Details
	StatusCode int
	SubCode    uuid.UUID
	cause      *error // TODO: read about interfaces, should this be a pointer? Or is an interface automatically nil-able?
}

type Details map[string]interface{}

var sentinel = errors.New("error did not have a cause")

func (e *Error) Error() string {
	detailsBytes, err := json.Marshal(e.Details)
	var detailsString string
	if err != nil {
		detailsString = ""
	} else {
		detailsString = string(detailsBytes)
	}
	return fmt.Sprintf(
		"%s (details: %s, status code: %d, sub code: %s, cause: %s)",
		e.Message, detailsString, e.StatusCode, e.SubCode.String(), e.Unwrap(),
	)
}

func (e *Error) Unwrap() error {
	if e.cause != nil {
		return *e.cause
	}
	return sentinel
}

func (e *Error) Logged(logger *zap.SugaredLogger) *Error {
	logDetails := make([]interface{}, len(e.Details)*2)
	idx := 0
	for k, v := range e.Details {
		logDetails[idx] = k
		idx++
		logDetails[idx] = v
		idx++
	}
	if e.cause != nil {
		logDetails = append(logDetails, "err")
		logDetails = append(logDetails, (*e.cause).Error())
	}
	logger.Errorw(e.Message, logDetails...)
	return e
}

func SetResponse(err error, c *gin.Context) {
	var e *Error
	if errors.As(err, &e) {
		var cause *string = nil
		if e.Unwrap() != sentinel {
			causeMsg := e.Unwrap().Error()
			cause = &causeMsg
		}
		c.JSON(
			e.StatusCode,
			gin.H{
				"message":  e.Message,
				"details":  e.Details,
				"sub_code": e.SubCode,
				"cause":    cause,
			},
		)
	} else {
		SetResponse(UnclassifiedError("unclassified error", Details{}, &err), c)
	}
}

func InvalidInput(message string, details Details, cause *error) *Error {
	return &Error{
		Message:    message,
		Details:    details,
		StatusCode: http.StatusBadRequest,
		SubCode:    uuid.MustParse("8fa6a458-07ad-40ec-a357-616c59ddb7ad"),
		cause:      cause,
	}
}

func ServiceNotFound(message string, details Details, cause *error) *Error {
	return &Error{
		Message:    message,
		Details:    details,
		StatusCode: http.StatusNotFound,
		SubCode:    uuid.MustParse("4b281f39-2eaf-4e09-8b6f-ffb277ea0cbb"),
		cause:      cause,
	}
}

func InitializationError(message string, details Details, cause *error) *Error {
	return &Error{
		Message:    message,
		Details:    details,
		StatusCode: http.StatusInternalServerError,
		SubCode:    uuid.MustParse("d4a0094d-bd7d-4471-871f-36b53c922044"),
		cause:      cause,
	}
}

func DatabaseError(message string, details Details, cause *error) *Error {
	return &Error{
		Message:    message,
		Details:    details,
		StatusCode: http.StatusInternalServerError,
		SubCode:    uuid.MustParse("f4bb1d18-f4ca-4401-9a2a-8e201e707d5a"),
		cause:      cause,
	}
}

func UnclassifiedError(message string, details Details, cause *error) *Error {
	return &Error{
		Message:    message,
		Details:    details,
		StatusCode: http.StatusInternalServerError,
		SubCode:    uuid.MustParse("4faf26fb-3996-4746-98ca-484fb27ffb23"),
		cause:      cause,
	}
}
