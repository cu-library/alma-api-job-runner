// Copyright 2020 Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestAlmaJobMarshalandUnmarshal(t *testing.T) {

	expectedMarshalled := `<job link="joblink">
  <id>M26714670000011</id>
  <name>Export Physical Items (name)</name>
  <description>Export Physical Items (description)</description>
  <type desc="Manual">MANUAL</type>
  <category desc="Normalization">NORMALIZATION</category>
  <content desc="All Titles">BIB_MMS</content>
  <schedule desc="Not scheduled">NOT_SCHEDULED</schedule>
  <creator>A Cool Cat</creator>
  <next_run>2015-07-20</next_run>
  <parameters>
    <parameter>
      <name desc="param desc">DESC</name>
      <value>AlphaBetaGamma</value>
    </parameter>
    <parameter>
      <name desc="param desc 2">DESC2</name>
      <value>AlphaBetaGammaDelta</value>
    </parameter>
  </parameters>
  <related_profile link="related profile link">ID123</related_profile>
  <additional_info link="additional info link">additional info value</additional_info>
</job>`

	expectedUnmarshalled := &AlmaJob{
		XMLName: xml.Name{
			Space: "",
			Local: "job",
		},
		Link:        "joblink",
		ID:          "M26714670000011",
		Name:        "Export Physical Items (name)",
		Description: "Export Physical Items (description)",
		Type: &DescAndValue{
			Desc:  "Manual",
			Value: "MANUAL",
		},
		Category: &DescAndValue{
			Desc:  "Normalization",
			Value: "NORMALIZATION",
		},
		Content: &DescAndValue{
			Desc:  "All Titles",
			Value: "BIB_MMS",
		},
		Schedule: &DescAndValue{
			Desc:  "Not scheduled",
			Value: "NOT_SCHEDULED",
		},
		Creator: "A Cool Cat",
		NextRun: "2015-07-20",
		Parameters: []Parameter{
			{
				Name: DescAndValue{
					Desc:  "param desc",
					Value: "DESC",
				},
				Value: "AlphaBetaGamma",
			},
			{
				Name: DescAndValue{
					Desc:  "param desc 2",
					Value: "DESC2",
				},
				Value: "AlphaBetaGammaDelta",
			},
		},
		RelatedProfile: &LinkAndValue{
			Link:  "related profile link",
			Value: "ID123",
		},
		AdditionalInfo: &LinkAndValue{
			Link:  "additional info link",
			Value: "additional info value",
		},
	}

	unmarshalled := &AlmaJob{}
	err := xml.Unmarshal([]byte(expectedMarshalled), unmarshalled)
	if err != nil {
		t.Fatal(err)
	}

	marshalled, err := xml.MarshalIndent(unmarshalled, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(expectedMarshalled, string(marshalled)) == false {
		t.Logf("\n%#v\n%#v\n", expectedMarshalled, string(marshalled))
		t.Fatal("Expected marshalled and marshalled Alma Job are not equal.")
	}

	if reflect.DeepEqual(expectedUnmarshalled, unmarshalled) == false {
		t.Logf("\n%#v\n%#v\n", expectedUnmarshalled, unmarshalled)
		t.Fatal("Expected unmarshalled and unmarshalled Alma Job are not equal.")
	}
}

func TestAlmaJobInstanceMarshalandUnmarshal(t *testing.T) {

	expectedMarshalled := `<job_instance link="job instance link">
  <id>1108569450000121</id>
  <external_id>1108569450000122</external_id>
  <name>Export Physical Items - war and love - 04/15/2015 16:07</name>
  <submitted_by desc="submitted by exl">exl_impl</submitted_by>
  <submit_time>2015-04-15T13:07:40.800Z</submit_time>
  <start_time>2015-04-15T13:07:40.868Z</start_time>
  <end_time>2015-04-15T13:07:44.359Z</end_time>
  <progress>50.776</progress>
  <status desc="status desc queued">QUEUED</status>
  <status_date>2015-07-20</status_date>
  <alerts>
    <alert desc="alert desc">alert_general_success</alert>
  </alerts>
  <counters>
    <counter>
      <type desc="counter extreme">4</type>
      <value>It&#39;s at 4!</value>
    </counter>
  </counters>
  <actions>
    <action>action 1</action>
    <action>action 2</action>
  </actions>
  <job_info link="job link">
    <id>1108569450000121</id>
    <name>Export Physical Items - war and love - 04/15/2015 16:07</name>
    <description>job description</description>
    <type desc="create set">CREATE_SET</type>
    <category desc="normalization">NORMALIZATION</category>
  </job_info>
</job_instance>`

	expectedUnmarshalled := &AlmaJobInstance{
		XMLName: xml.Name{
			Space: "",
			Local: "job_instance",
		},
		Link:       "job instance link",
		ID:         "1108569450000121",
		ExternalID: "1108569450000122",
		Name:       "Export Physical Items - war and love - 04/15/2015 16:07",
		SubmittedBy: &DescAndValue{
			Desc:  "submitted by exl",
			Value: "exl_impl",
		},
		SubmitTime: "2015-04-15T13:07:40.800Z",
		StartTime:  "2015-04-15T13:07:40.868Z",
		EndTime:    "2015-04-15T13:07:44.359Z",
		Progress:   50.776,
		Status: &DescAndValue{
			Desc:  "status desc queued",
			Value: "QUEUED",
		},
		StatusDate: "2015-07-20",
		Alerts: []DescAndValue{
			{
				Desc:  "alert desc",
				Value: "alert_general_success",
			},
		},
		Counters: []Counter{
			{
				Type: DescAndValue{
					Desc:  "counter extreme",
					Value: "4",
				},
				Value: "It's at 4!",
			},
		},
		Actions: []string{
			"action 1",
			"action 2",
		},
		JobInfo: &AlmaJobInfo{
			Link:        "job link",
			ID:          "1108569450000121",
			Name:        "Export Physical Items - war and love - 04/15/2015 16:07",
			Description: "job description",
			Type: &DescAndValue{
				Desc:  "create set",
				Value: "CREATE_SET",
			},
			Category: &DescAndValue{
				Desc:  "normalization",
				Value: "NORMALIZATION",
			},
		},
	}

	unmarshalled := &AlmaJobInstance{}
	err := xml.Unmarshal([]byte(expectedMarshalled), unmarshalled)
	if err != nil {
		t.Fatal(err)
	}

	marshalled, err := xml.MarshalIndent(unmarshalled, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(expectedMarshalled, string(marshalled)) == false {
		t.Logf("\n%#v\n%#v\n", expectedMarshalled, string(marshalled))
		t.Fatal("Expected marshalled and marshalled Alma Job Instances are not equal.")
	}

	if reflect.DeepEqual(expectedUnmarshalled, unmarshalled) == false {
		t.Logf("\n%#v\n%#v\n", expectedUnmarshalled, unmarshalled)
		t.Fatal("Expected unmarshalled and unmarshalled Alma Job Instances are not equal.")
	}
}
