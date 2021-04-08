// Copyright 2021 Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestLoadParameters(t *testing.T) {
	expectedJob := AlmaJob{
		XMLName: xml.Name{
			Space: "",
			Local: "job",
		},
		Parameters: []Parameter{
			{
				Name: DescAndValue{
					Value: "task_MmsTaggingParams_boolean",
				},
				Value: "NONE",
			},
			{
				Name: DescAndValue{
					Value: "set_id",
				},
				Value: "4000000000000",
			},
			{
				Name: DescAndValue{
					Value: "job_name",
				},
				Value: "A Job Name Here",
			},
		},
	}

	content := []byte(`<job>
	<parameters>
		<parameter>
			<name>task_MmsTaggingParams_boolean</name>
			<value>NONE</value>
		</parameter>
		<parameter>
			<name>set_id</name>
			<value>4000000000000</value>
		</parameter>
		<parameter>
			<name>job_name</name>
			<value>A Job Name Here</value>
		</parameter>
	</parameters>
</job>`)

	tmpFile, err := ioutil.TempFile("", "*.params.xml")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up.

	_, err = tmpFile.Write(content)
	if err != nil {
		tmpFile.Close()
		t.Error(err)
	}

	params, err := LoadParameters(tmpFile.Name())
	if err != nil {
		t.Error(err)
	}

	if reflect.DeepEqual(expectedJob, params) == false {
		t.Logf("\n%#v\n%#v\n", expectedJob, params)
		t.Fatal("Expected job and job loaded through LoadParameters() are not equal.")
	}
}
