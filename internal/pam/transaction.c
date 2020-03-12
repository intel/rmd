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

#include "_cgo_export.h"
#include <security/pam_appl.h>
#include <string.h>
#include <stdio.h>

int stringLength(char *str) {
    int i=0;
    while (str[i] != '\0') {
        i++;
    }
    return i;
}

int cb_pam_conv(
	int num_msg,
	const struct pam_message **msg,
	struct pam_response **resp,
	void *appdata_ptr)
{
	*resp = (struct pam_response *) calloc(num_msg, sizeof **resp);
	if (num_msg <= 0 || num_msg > PAM_MAX_NUM_MSG) {
		return PAM_CONV_ERR;
	}
	if (!*resp) {
		return PAM_BUF_ERR;
	}
	for (size_t i = 0; i < num_msg; ++i) {
		// cbPAMConv is a Go Wrapper to prompt the Go application for password.
		// return values are captured in cbConv_return as the Go Wrapper exports cbConv.
		struct cbConv_return result = cbConv(
				msg[i]->msg_style,
				(char *)msg[i]->msg,
				(long)appdata_ptr);
		if (result.r1 != PAM_SUCCESS) {
			for (size_t i = 0; i < num_msg; ++i) {
				if ((*resp)[i].resp) {
					memset((*resp)[i].resp, 0, stringLength((*resp)[i].resp));
					free((*resp)[i].resp);
				}
			}
			memset(*resp, 0, num_msg * sizeof *resp);
			free(*resp);
			*resp = NULL;
			return PAM_CONV_ERR;
		}
		(*resp)[i].resp = result.r0;
	}
	return PAM_SUCCESS;
}

void init_pam_conv(struct pam_conv *conv, long c)
{
	conv->conv = cb_pam_conv;
	conv->appdata_ptr = (void *) c;
}
