// Copyright 2023 Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/xml"
	"errors"
	"fmt"
)

// ErrAPIError is an error which is used when the Alma API returns an error.
var ErrAPIError = errors.New("an API error occurred")

// APIError is a struct which stores the data from Alma API errors.
type APIError struct {
	XMLName   xml.Name `xml:"web_service_result"`
	ErrorList []struct {
		Error struct {
			ErrorCode    string `xml:"errorCode"`
			ErrorMessage string `xml:"errorMessage"`
		} `xml:"error"`
	} `xml:"errorList"`
}

// Collapse is a method on APIErrors which returns the first APIError as
// a Go error. This was done this way because the API returns different
// error codes which mean the same thing.
func (e *APIError) Collapse() error {
	if len(e.ErrorList) > 0 {
		return fmt.Errorf("%w: %v - %v", ErrAPIError, e.ErrorList[0].Error.ErrorCode, e.ErrorList[0].Error.ErrorMessage)
	}
	return fmt.Errorf("%w: unknown error", ErrAPIError)
}
