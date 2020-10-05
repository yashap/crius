package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Error struct {
	Message    string
	StatusCode int
	SubCode    uuid.UUID
	cause      *error // TODO: read about interfaces, should this be a pointer? Or is an interface automatically nil-able?
}

var sentinel error = errors.New("error did not have a cause")

func (e *Error) Error() string {
	return fmt.Sprintf(
		"%s (status code: %d, sub code: %s, cause: %s)",
		e.Message, e.StatusCode, e.SubCode.String(), e.Unwrap(),
	)
}

func (e *Error) Unwrap() error {
	if e.cause != nil {
		return *e.cause
	}
	return sentinel
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
				"sub_code": e.SubCode.String(),
				"cause":    cause,
			},
		)
	} else {
		SetResponse(UnclassifiedError("unclassified error", &err), c)
	}
}

func InvalidInput(message string, cause *error) error {
	return &Error{
		Message:    message,
		StatusCode: http.StatusBadRequest,
		SubCode:    uuid.MustParse("8fa6a458-07ad-40ec-a357-616c59ddb7ad"),
		cause:      cause,
	}
}

func ServiceNotFound(message string, cause *error) error {
	return &Error{
		Message:    message,
		StatusCode: http.StatusNotFound,
		SubCode:    uuid.MustParse("4b281f39-2eaf-4e09-8b6f-ffb277ea0cbb"),
		cause:      cause,
	}
}

func EndpointNotFound(message string, cause *error) error {
	return &Error{
		Message:    message,
		StatusCode: http.StatusNotFound,
		SubCode:    uuid.MustParse("0e86e5ad-e332-4962-b138-34dddade1dd1"),
		cause:      cause,
	}
}

func DatabaseError(message string, cause *error) error {
	return &Error{
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		SubCode:    uuid.MustParse("f4bb1d18-f4ca-4401-9a2a-8e201e707d5a"),
		cause:      cause,
	}
}

func UnclassifiedError(message string, cause *error) error {
	return &Error{
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		SubCode:    uuid.MustParse("4faf26fb-3996-4746-98ca-484fb27ffb23"),
		cause:      cause,
	}
}
