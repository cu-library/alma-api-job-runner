// Copyright 2020 Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/xml"
)

// AlmaJob is a type which maps XML data from the API about jobs to Go structs.
// https://developers.exlibrisgroup.com/alma/apis/docs/xsd/rest_job.xsd/
type AlmaJob struct {
	XMLName        xml.Name      `xml:"job"`
	Link           string        `xml:"link,attr,omitempty"`
	ID             string        `xml:"id,omitempty"`
	Name           string        `xml:"name,omitempty"`
	Description    string        `xml:"description,omitempty"`
	Type           *DescAndValue `xml:"type,omitempty"`
	Category       *DescAndValue `xml:"category,omitempty"`
	Content        *DescAndValue `xml:"content,omitempty"`
	Schedule       *DescAndValue `xml:"schedule,omitempty"`
	Creator        string        `xml:"creator,omitempty"`
	NextRun        string        `xml:"next_run,omitempty"`
	Parameters     []Parameter   `xml:"parameters>parameter,omitempty"`
	RelatedProfile *LinkAndValue `xml:"related_profile,omitempty"`
	AdditionalInfo *LinkAndValue `xml:"additional_info,omitempty"`
}

// AlmaJobInstance is a type which maps XML data from the API about job instances to Go structs.
// https://developers.exlibrisgroup.com/alma/apis/docs/xsd/rest_job_instance.xsd
type AlmaJobInstance struct {
	XMLName     xml.Name       `xml:"job_instance"`
	Link        string         `xml:"link,attr,omitempty"`
	ID          string         `xml:"id,omitempty"`
	ExternalID  string         `xml:"external_id,omitempty"`
	Name        string         `xml:"name,omitempty"`
	SubmittedBy *DescAndValue  `xml:"submitted_by,omitempty"`
	SubmitTime  string         `xml:"submit_time,omitempty"`
	StartTime   string         `xml:"start_time,omitempty"`
	EndTime     string         `xml:"end_time,omitempty"`
	Progress    float64        `xml:"progress,omitempty"`
	Status      *DescAndValue  `xml:"status,omitempty"`
	StatusDate  string         `xml:"status_date,omitempty"`
	Alerts      []DescAndValue `xml:"alerts>alert,omitempty"`
	Counters    []Counter      `xml:"counters>counter,omitempty"`
	Actions     []string       `xml:"actions>action,omitempty"`
	JobInfo     *AlmaJobInfo   `xml:"job_info"`
}

// AlmaJobInfo is a type which stores info about a job.
type AlmaJobInfo struct {
	Link        string        `xml:"link,attr,omitempty"`
	ID          string        `xml:"id,omitempty"`
	Name        string        `xml:"name,omitempty"`
	Description string        `xml:"description,omitempty"`
	Type        *DescAndValue `xml:"type,omitempty"`
	Category    *DescAndValue `xml:"category,omitempty"`
}

// DescAndValue stores the value and the desc attribute
// of an element.
type DescAndValue struct {
	Desc  string `xml:"desc,attr,omitempty"`
	Value string `xml:",chardata"`
}

// LinkAndValue stores the value and the link attribute
// of an element.
type LinkAndValue struct {
	Link  string `xml:"link,attr,omitempty"`
	Value string `xml:",chardata"`
}

// Parameter stores a job's related parameters.
type Parameter struct {
	Name  DescAndValue `xml:"name,omitempty"`
	Value string       `xml:"value,omitempty"`
}

// Counter stores a job instance counter.
type Counter struct {
	Type  DescAndValue `xml:"type,omitempty"`
	Value string       `xml:"value,omitempty"`
}
