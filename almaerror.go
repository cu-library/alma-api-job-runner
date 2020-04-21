// Copyright 2020 Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/xml"
	"errors"
	"fmt"
)

// APIError is a struct which stores the data from Alma API errors.
type APIError struct {
	XMLName   xml.Name `xml:"web_service_result"`
	ErrorList []struct {
		Error struct {
			ErrorCode int `xml:"errorCode"`
		} `xml:"error"`
	} `xml:"errorList"`
}

// Collapse is a method on APIErrors which returns the first APIError as
// a Go error. This was done this way because the API returns different
// error codes which mean the same thing.
func (e *APIError) Collapse() error {
	for _, ele := range e.ErrorList {
		switch ele.Error.ErrorCode {
		case 402215:
			return fmt.Errorf("invalid job id format (%v)", ele.Error.ErrorCode)
		case 402216:
			return fmt.Errorf("invalid job id (%v)", ele.Error.ErrorCode)
		case 402218:
			return fmt.Errorf("invalid job instance id (%v)", ele.Error.ErrorCode)
		case 402220:
			return fmt.Errorf("operation was not provided (%v)", ele.Error.ErrorCode)
		case 402221:
			return fmt.Errorf("operation is not supported (%v)", ele.Error.ErrorCode)
		case 402222:
			return fmt.Errorf("execution threshold reached (%v)", ele.Error.ErrorCode)
		case 402223:
			return fmt.Errorf("execution threshold reached (%v)", ele.Error.ErrorCode)
		case 402224:
			return fmt.Errorf("an internal error occured (%v)", ele.Error.ErrorCode)
		case 402225:
			return fmt.Errorf("an internal error occured (%v)", ele.Error.ErrorCode)
		case 402226:
			return fmt.Errorf("an internal error occured (%v)", ele.Error.ErrorCode)
		case 402228:
			return fmt.Errorf("mandatory parameter is missing from input (%v)", ele.Error.ErrorCode)
		case 402229:
			return fmt.Errorf("mandatory parameter value is empty (%v)", ele.Error.ErrorCode)
		case 402248:
			return fmt.Errorf("cannot submit scheduled job (%v)", ele.Error.ErrorCode)
		case 402249:
			return fmt.Errorf("invalid scheduled job category (%v)", ele.Error.ErrorCode)
		case 402231:
			return fmt.Errorf("job in consisted of more than one task - executing such job is currently not supported via the API (%v)", ele.Error.ErrorCode)
		default:
			return errors.New("unknown error")
		}
	}
	return errors.New("unreachable error")
}
