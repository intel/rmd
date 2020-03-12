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

package pam

import "sync"

var cb struct {
	sync.Mutex
	m map[int]interface{}
	c int
}

func init() {
	cb.m = make(map[int]interface{})
}

func cbAdd(v interface{}) int {
	cb.Lock()
	defer cb.Unlock()
	cb.c++
	cb.m[cb.c] = v
	return cb.c
}

func cbGet(c int) interface{} {
	cb.Lock()
	defer cb.Unlock()
	if v, ok := cb.m[c]; ok {
		return v
	}
	panic("Callback pointer not found")
}

func cbDelete(c int) {
	cb.Lock()
	defer cb.Unlock()
	if _, ok := cb.m[c]; !ok {
		panic("Callback pointer not found")
	}
	delete(cb.m, c)
}
