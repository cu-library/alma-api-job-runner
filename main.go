// Copyright 2020 Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cu-library/overridefromenv"
)

// A version flag, which should be overwritten when building using ldflags.
var version = "devel"

const (
	// EnvPrefix is the prefix for environment variables which override unset flags.
	EnvPrefix string = "ALMA_API_JOB_RUNNER_"

	// DefaultSMTPPort is the default port to use when connecting to the SMTP server.
	DefaultSMTPPort = 25
)

func main() {
	// Define the command line flags.
	almaAPIServer := flag.String("almaapi", "", "The Alma API server URL to use. Required. (ex: api-ca.hosted.exlibrisgroup.com)")
	almaAPIKey := flag.String("almaapikey", "", "The Alma API key. Required.")
	url := flag.String("url", "", "The URL to which the job's parameters should be POST'd. Required.")
	xmlParams := flag.String("paramsxml", "", "The filepath of the XML representation of the job's parameters. Required.")
	sendEmail := flag.Bool("email", true, "Send an email report.")
	smtpServer := flag.String("smtpserver", "", "The SMTP server to use for sending report emails.")
	smtpPort := flag.Int("smtpport", DefaultSMTPPort, "The port to use when connecting to the SMTP server.")
	mailTo := flag.String("mailto", "", "The email address to send reports to, comma delimited.")
	mailFrom := flag.String("mailfrom", "", "The email address reports are send from.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "alma-api-job-runner:\n")
		fmt.Fprintf(os.Stderr, "Run a manual job in Alma using the Jobs API.\n")
		fmt.Fprintf(os.Stderr, "Version %v\n", version)
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "  Environment variables read when flag is unset:")

		flag.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(os.Stderr, "  %v%v\n", EnvPrefix, strings.ToUpper(f.Name))
		})
	}

	// Process the flags.
	flag.Parse()

	// If any flags have not been set, see if there are
	// environment variables that set them.
	err := overridefromenv.Override(flag.CommandLine, EnvPrefix)
	if err != nil {
		log.Fatalln(err)
	}

	// Exit if any required flags are not set.
	if *almaAPIServer == "" {
		log.Fatal("FATAL: An Alma API Server URL is required. https://developers.exlibrisgroup.com/alma/apis/#calling")
	}
	if *almaAPIKey == "" {
		log.Fatal("FATAL: An Alma API Key is required. https://developers.exlibrisgroup.com/alma/apis/#defining")
	}
	if *url == "" {
		log.Fatal("FATAL: A URL is required. https://developers.exlibrisgroup.com/blog/Working-with-the-Alma-Jobs-API/")
	}
	if *xmlParams == "" {
		log.Fatal("FATAL: An XML file of the job's parameters is required. https://developers.exlibrisgroup.com/blog/Working-with-the-Alma-Jobs-API/")
	}
	if *sendEmail {
		if *smtpServer == "" {
			log.Fatal("FATAL: A SMTP server is required if the email option is being used.")
		}
		if *mailTo == "" {
			log.Fatal("FATAL: At least one email address to send reports to must be provided if using the email option.")
		}
		if *mailFrom == "" {
			log.Fatal("FATAL: An email address reports are send from must be provided if using the email option.")
		}
	}

	now := time.Now()

	// Create a buffer to store the email message.
	// The email report is a copy of the log messages.
	emailMessage := new(bytes.Buffer)

	// Split log output to both stderr and the email message.
	log.SetOutput(io.MultiWriter(os.Stderr, emailMessage))

	log.Println("Using alma-api-job-runner", version)
	log.Println("Flags:")
	log.Println(now)
	log.Println(smtpPort)
}
