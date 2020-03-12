// SPDX short identifier: BSD-2-Clause
// Copyright 2011, krockot
// Copyright 2015, Michael Steinert <mike.steinert@gmail.com>
// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:

// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.

// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package pam provides a wrapper for the PAM application API.
package pam

//#include <security/pam_appl.h>
//#include <stdlib.h>
//#cgo CFLAGS: -Wall -std=c99
//#cgo LDFLAGS: -lpam
//void init_pam_conv(struct pam_conv *convfunc, long callbackindex);
import "C"

import (
	"unsafe"
)

// ConvResponse is the type of message that the conversation handler should display.
type ConvResponse int

// Conversation handler style types.
const (
	// PromptEchoOff indicates the conversation handler should obtain a
	// string without echoing any text.
	PromptEchoOff ConvResponse = C.PAM_PROMPT_ECHO_OFF
	// PromptEchoOn indicates the conversation handler should obtain a
	// string while echoing text.
	PromptEchoOn = C.PAM_PROMPT_ECHO_ON
	// ErrorMsg indicates the conversation handler should display an
	// error message.
	ErrorMsg = C.PAM_ERROR_MSG
	// TextInfo indicates the conversation handler should display some
	// text.
	TextInfo = C.PAM_TEXT_INFO
)

// ConversationHandler is an interface for objects that can be used as
// conversation callbacks during PAM authentication.
type ConversationHandler interface {
	// PromptPassword receives a message ConvResponse and a message string. If the
	// message ConvResponse is PromptEchoOff or PromptEchoOn then the function
	// should return a response string.
	PromptPassword(ConvResponse, string) (string, error)
}

// ConversationFunc is an adapter to allow the use of ordinary functions as
// conversation callbacks.
type ConversationFunc func(ConvResponse, string) (string, error)

// PromptPassword is a conversation callback adapter.
func (f ConversationFunc) PromptPassword(flag ConvResponse, msg string) (string, error) {
	return f(flag, msg)
}

// cbConv is a GO wrapper for the conversation callback function.
// NOTE: DO NOT REMOVE BELOW COMMENT,it's used as struct type to capture values in caller func.
//export cbConv
func cbConv(s C.int, msg *C.char, callbackindex int) (*C.char, C.int) {
	var r string
	var err error
	v := cbGet(callbackindex)
	switch cb := v.(type) {
	case ConversationHandler:
		r, err = cb.PromptPassword(ConvResponse(s), C.GoString(msg))
	}
	if err != nil {
		return nil, C.PAM_CONV_ERR
	}
	return C.CString(r), C.PAM_SUCCESS
}

// Transaction is the application's handle for a PAM transaction.
type Transaction struct {
	handle        *C.pam_handle_t
	convfunc      *C.struct_pam_conv
	response      C.int
	callbackindex int
}

// StartTransaction initiates a new PAM transaction.
// Returned transaction provides an interface to the remainder of the API.
func StartTransaction(service, user string, handler ConversationHandler) (*Transaction, error) {
	t := &Transaction{
		convfunc:      &C.struct_pam_conv{},
		callbackindex: cbAdd(handler),
	}
	C.init_pam_conv(t.convfunc, C.long(t.callbackindex))
	var s *C.char
	s = C.CString(service)
	defer C.free(unsafe.Pointer(s))
	var u *C.char
	if len(user) != 0 {
		u = C.CString(user)
		defer C.free(unsafe.Pointer(u))
	}
	// Initiate a transaction with the conversation func ptr.
	t.response = C.pam_start(s, u, t.convfunc, &t.handle)
	if t.response != C.PAM_SUCCESS {
		return nil, t
	}
	return t, nil
}

// StartFunc registers the handler func as a conversation handler.
func StartFunc(service, user string, handler func(ConvResponse, string) (string, error)) (*Transaction, error) {
	return StartTransaction(service, user, ConversationFunc(handler))
}

// Flags are inputs to various PAM functions than be combined with a bitwise
// or. Refer to the official PAM documentation for which flags are accepted
// by which functions.
type Flags int

// PAM Flag types.
const (
	// No Flags.
	NoFlag Flags = 0
	// Silent indicates that no messages should be emitted.
	Silent = C.PAM_SILENT
	// DisallowNullAuthtok indicates that authorization should fail
	// if the user does not have a registered authentication token.
	DisallowNullAuthtok = C.PAM_DISALLOW_NULL_AUTHTOK
	// EstablishCred indicates that credentials should be established
	// for the user.
	EstablishCred = C.PAM_ESTABLISH_CRED
	// DeleteCred inidicates that credentials should be deleted.
	DeleteCred = C.PAM_DELETE_CRED
	// ReinitializeCred indicates that credentials should be fully
	// reinitialized.
	ReinitializeCred = C.PAM_REINITIALIZE_CRED
	// RefreshCred indicates that the lifetime of existing credentials
	// should be extended.
	RefreshCred = C.PAM_REFRESH_CRED
	// ChangeExpiredAuthtok indicates that the authentication token
	// should be changed if it has expired.
	ChangeExpiredAuthtok = C.PAM_CHANGE_EXPIRED_AUTHTOK
)

// Authenticate is used to authenticate the user.
// Valid flags: Silent, DisallowNullAuthtok
func (t *Transaction) Authenticate(f Flags) error {
	t.response = C.pam_authenticate(t.handle, C.int(f))
	if t.response != C.PAM_SUCCESS {
		return t
	}
	return nil
}

// EndTransaction cleans up the PAM handle and deletes the callback
func EndTransaction(t *Transaction) {
	C.pam_end(t.handle, t.response)
	cbDelete(t.callbackindex)
}

// Called when returning a transaction object for the return type as error.
func (t *Transaction) Error() string {
	return C.GoString(C.pam_strerror(t.handle, C.int(t.response)))
}
