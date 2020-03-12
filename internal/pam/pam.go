package pam

import (
	"errors"
	"github.com/intel/rmd/internal/pam/config"
	log "github.com/sirupsen/logrus"
)

// Credential represents user provided credential
type Credential struct {
	Username string
	Password string
}

// PAMResponseHandler handles the communication between PAM client and PAM module
func (c Credential) PAMResponseHandler(s ConvResponse, msg string) (string, error) {
	switch s {
	case PromptEchoOff:
		return c.Password, nil
	case PromptEchoOn:
		log.Info(msg)
		return c.Password, nil
	case ErrorMsg:
		log.Error(msg)
		return "", nil
	case TextInfo:
		log.Info(msg)
		return "", nil
	default:
		return "", errors.New("unrecognized conversation message style")
	}
}

// pamTxAuthenticate authenticates a PAM transaction
func pamTxAuthenticate(transaction *Transaction) error {
	err := transaction.Authenticate(0)
	EndTransaction(transaction)
	return err
}

// PAMAuthenticate performs PAM authentication for the user credentials provided
func (c Credential) PAMAuthenticate() error {
	tx, err := c.PAMStartFunc()
	if err != nil {
		return err
	}
	err = pamTxAuthenticate(tx)
	return err
}

// pamStartFunc starts the conversation between PAM client and PAM module
func pamStartFunc(service string, user string, handler func(ConvResponse, string) (string, error)) (*Transaction, error) {
	tx, err := StartFunc(service, user, handler)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// PAMStartFunc establishes the connection to PAM module
func (c Credential) PAMStartFunc() (*Transaction, error) {
	return pamStartFunc(config.GetPAMConfig().Service, c.Username, c.PAMResponseHandler)
}
