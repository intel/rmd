package pam

import (
	"testing"
)

func TestPAMStartFunc(t *testing.T) {
	_, err := pamStartFunc("", "", nil)
	if err != nil {
		t.Error(err)
	}
}

func TestPAMTxAuthenticate(t *testing.T) {
	// valid credential
	c := Credential{"user", "user1"}

	// valid service name
	service := "rmd"

	tx, err := pamStartFunc(service, c.Username, c.PAMResponseHandler)
	if err != nil {
		t.Fatal(err)
	}

	err = pamTxAuthenticate(tx)
	if err != nil {
		t.Error(err)
	}
}

func TestPAMAuthenticate(t *testing.T) {

	// Litmus test start func
	_, err := pamStartFunc("", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Please use credentials different from those defined in Berkeley db for a consistent error message
	testCases := []struct {
		username      string
		password      string
		description   string
		desiredResult string
	}{
		{"user", "user1", "Valid Berkeley db user", ""},
		{"x", "y", "Invalid Berkeley db user", "User not known to the underlying authentication module"},
		{"user", "user", "Incorrect Berkeley db user", "Authentication failure"},
		// Edit unix credentials here according to your testing platform
		// {"root", "s", "Valid unix user", ""},
		{"a", "b", "Invalid unix user", "User not known to the underlying authentication module"},
		{"root", "s1", "Incorrect unix user", "User not known to the underlying authentication module"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			c := Credential{testCase.username, testCase.password}
			err := c.PAMAuthenticate()
			if testCase.desiredResult == "" {
				if err != nil {
					t.Error(err)
				}
			} else {
				if err == nil {
					t.Error("No error detected as desired. Please check test inputs")
				}
				if err.Error() != testCase.desiredResult {
					t.Error(err)
				}
			}
		})
	}
}
