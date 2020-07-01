// Copyright 2020 Alexander Eldeib
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
// OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
// CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	mrand "math/rand"
)

const numeric = "0123456789"
const lower = "abcdefghijklmnopqrstuvwxyz"
const upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const safeBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const safeLowerBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

// https://github.com/kubernetes-sigs/cluster-api-provider-azure/blob/60b7c6058550ae694935fb03103460a2efa4e332/pkg/cloud/azure/services/virtualmachines/virtualmachines.go#L215
// GenerateRandomString generates a random string of lenth n, panicking on failure.
func GenerateRandomString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		fmt.Printf("error in generate random: %+#v", err.Error())
	}
	return base64.StdEncoding.EncodeToString(b) //, err
}

func RandomLowercaseString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = safeLowerBytes[mrand.Intn(len(safeLowerBytes))]
	}
	return string(b)
}

func GenerateSafeRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = safeBytes[mrand.Intn(len(safeBytes))]
	}
	return string(b)
}

func RandomLowerAlpha(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = lower[mrand.Intn(len(lower))]
	}
	return string(b)
}
