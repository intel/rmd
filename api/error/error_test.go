package error

import (
	"testing"

	"errors"
	"fmt"
	"net/http"
)

func TestAppError(t *testing.T) {
	err := func() error {
		return errors.New("Return a error just for test")
	}
	// In a model, after we catch an error, reform it and return AppError.
	appErr := AppError{http.StatusNotFound, "Reason Message", err()}
	fmt.Printf("%v", appErr)
	fmt.Println(appErr.Error(), appErr.Code)

	ep := NewAppError(http.StatusNotFound, err(), "Reason Message")
	fmt.Println(ep.Err, ep.Code)
	// build from formated string
	e := AppErrorf(ep.Code, "Error for test, reason: %d\n", ep.Code)
	fmt.Printf("%T, %s\n", e, e.Error())
	// Output:
	// {404 Reason Message. Return a error just for test.}
	// Reason Message.. Details: Return a error just for test. 404
	// Return a error just for test. 404
	// *error.AppError, Error for test, reason: 404
}
