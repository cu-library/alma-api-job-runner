// Copyright 2020 Carleton University Library All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
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
	almaAPIDomain := flag.String("apidomain", "", "The domain of the Alma API server URL to use. Required. (ex: api-ca.hosted.exlibrisgroup.com)")
	almaAPIKey := flag.String("apikey", "", "The Alma API key. Required.")
	destURL := flag.String("url", "", "The URL to which the job's parameters should be POST'd. Required.")
	params := flag.String("params", "", "A file storing the XML representation of the job's parameters. Required.")
	timeout := flag.Int("timeout", 10, "The number of seconds to wait on the Alma API when submitting requests.")
	maxRetries := flag.Int("retries", 5, "If calling the Alma API results in an error, how many times will the job be resubmitted.")
	sendEmail := flag.Bool("email", false, "Send an email report.")
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
	if *almaAPIDomain == "" {
		log.Fatal("FATAL: An Alma API Server domain is required. https://developers.exlibrisgroup.com/alma/apis/#calling")
	}
	if *almaAPIKey == "" {
		log.Fatal("FATAL: An Alma API Key is required. https://developers.exlibrisgroup.com/alma/apis/#defining")
	}
	if *destURL == "" {
		log.Fatal("FATAL: A URL is required. https://developers.exlibrisgroup.com/blog/Working-with-the-Alma-Jobs-API/")
	}
	if *params == "" {
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

	// Create a buffer to store the email message.
	// The email report is a copy of the log messages.
	emailMessage := new(bytes.Buffer)

	// Split log output to both stderr and the email message.
	log.SetOutput(io.MultiWriter(os.Stderr, emailMessage))

	// Add the arguments to the output for later debugging.
	log.Println("Using alma-api-job-runner version", version)
	log.Println("Running at: ", time.Now().Format(time.RFC3339))
	log.Println("Alma API Server (apidomain):", *almaAPIDomain)
	log.Println("Job URL (url):", *destURL)
	log.Println("Parameters file (params):", *params)

	// Build the request to the Alma API.
	completeURL, err := url.Parse(fmt.Sprintf("https://%v%v", *almaAPIDomain, *destURL))
	if err != nil {
		log.Println("Error building final url from arguments: ", err)
		if *sendEmail {
			err := SendEmail("alma-api-job-runner - error", emailMessage, smtpServer, smtpPort, mailTo, mailFrom)
			if err != nil {
				log.Println(err)
			}
		}
		os.Exit(1)
	}

	// Load the parameters XML file.
	// This is done to check that the XML is well formed and valid.
	loadedParams, err := LoadParameters(*params)
	if err != nil {
		log.Println("Error loading parameters: ", err)
		if *sendEmail {
			err := SendEmail("alma-api-job-runner - error", emailMessage, smtpServer, smtpPort, mailTo, mailFrom)
			if err != nil {
				log.Println(err)
			}
		}
		os.Exit(1)
	}

	// Retry for max retries
	for retry := 0; retry < *maxRetries; retry++ {
		jobInstanceID, err := SubmitJob(completeURL, *timeout, *almaAPIKey, loadedParams)
		if err != nil {
			log.Printf("Failed to submit job: %v, retrying. (%v/%v)\n", err, retry+1, maxRetries)
			continue
		}
		fmt.Println(jobInstanceID)
		if *sendEmail {
			err := SendEmail("alma-api-job-runner - success", emailMessage, smtpServer, smtpPort, mailTo, mailFrom)
			if err != nil {
				log.Println(err)
			}
		}
		os.Exit(0)
	}

	if *sendEmail {
		err := SendEmail("alma-api-job-runner - error", emailMessage, smtpServer, smtpPort, mailTo, mailFrom)
		if err != nil {
			log.Println(err)
		}
	}
	os.Exit(1)
}

// LoadParameters reads and unmarshals the contents of the params file.
func LoadParameters(params string) (loadedParams AlmaJob, err error) {
	paramsFilePath, err := filepath.Abs(params)
	if err != nil {
		return loadedParams, err
	}
	paramsFile, err := os.Open(paramsFilePath)
	if err != nil {
		return loadedParams, err
	}
	defer paramsFile.Close()
	decoder := xml.NewDecoder(paramsFile)
	err = decoder.Decode(&loadedParams)
	if err != nil {
		return loadedParams, err
	}
	return loadedParams, nil
}

// SubmitJob sends a POST HTTP message to the Alma API to execute the job.
func SubmitJob(url *url.URL, timeout int, almaAPIKey string, params AlmaJob) (jobInstanceID string, err error) {

	// Setup the job parameter data as a io.Reader
	marshaledParams := new(bytes.Buffer)
	encoder := xml.NewEncoder(marshaledParams)
	err = encoder.Encode(params)
	if err != nil {
		return "", err
	}

	// Setup the HTTP request
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	request, err := http.NewRequest("POST", url.String(), marshaledParams)
	if err != nil {
		return "", err
	}
	request.Header.Add("Authorization", "apikey "+almaAPIKey)

	// Do the request.
	// On error, drain and close the response body.
	resp, err := client.Do(request)
	if err != nil {
		if resp != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
		return "", err
	}

	// Log the remaning number of API calls.
	remainingCalls := resp.Header.Get("X-Exl-Api-Remaining")
	if remainingCalls != "" {
		log.Printf("%v Alma API calls remaining.\n", remainingCalls)
	}

	// If the response was a 400 error, we can parse the returned XML.
	if resp.StatusCode == 400 {
		apiErr := &APIError{}
		decoder := xml.NewDecoder(resp.Body)
		err := decoder.Decode(apiErr)
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("alma API request failed: %v couldn't read body: %v", resp.Status, err)
		}
		return "", apiErr.Collapse()
	}

	// If the Status != OK, there was an error we didn't catch yet.
	if resp.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("alma API request failed: %v couldn't read body: %v", resp.Status, err)
		}
		return "", fmt.Errorf("alma API request failed: %v - %v", resp.Status, string(bodyBytes))
	}

	returnedJob := &AlmaJob{}
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(returnedJob)
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	return path.Base(returnedJob.AdditionalInfo.Link), nil

}

// SendEmail sends an email using the provided configuration.
func SendEmail(subject string, emailMessage *bytes.Buffer, smtpServer *string, smtpPort *int, mailTo, mailFrom *string) error {
	to := TrimSpaceAll(strings.Split(*mailTo, ","))
	finalMsg := new(bytes.Buffer)
	finalMsg.WriteString(fmt.Sprintf("To: %v\r\n", strings.Join(to, ", ")))
	finalMsg.WriteString(fmt.Sprintf("Subject: %v\r\n", subject))
	finalMsg.WriteString("\r\n")
	bodyMsg := bytes.ReplaceAll(emailMessage.Bytes(), []byte("\n"), []byte("\r\n"))
	finalMsg.Write(bodyMsg)
	return smtp.SendMail(*smtpServer+":"+strconv.Itoa(*smtpPort), nil, *mailFrom, to, finalMsg.Bytes())
}

// TrimSpaceAll returns a version of trimMe where each element has been TrimSpace'd.
func TrimSpaceAll(trimMe []string) []string {
	var trimmed []string
	for _, cur := range trimMe {
		trimmed = append(trimmed, strings.TrimSpace(cur))
	}
	return trimmed
}
