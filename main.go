// Copyright 2021 Carleton University Library All rights reserved.
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
	name := flag.String("name", "Alma API Job Runner", "The name for the job, used for logging and reports only.")
	domain := flag.String("domain", "", "The domain of the Alma API server URL to use. Required. (ex: api-ca.hosted.exlibrisgroup.com)")
	key := flag.String("key", "", "The Alma API key. Required.")
	jobPath := flag.String("url", "", "The URL to which the job's parameters should be POST'd. Starts with a /. Required.")
	params := flag.String("params", "", "A file storing the XML representation of the job's parameters. Required.")
	timeout := flag.Int("timeout", 10, "The number of seconds to wait on the Alma API when submitting requests.")
	maxRetries := flag.Int("retries", 5, "If calling the Alma API results in an error, how many times will the job be resubmitted.")
	sendEmail := flag.Bool("email", false, "Send an email report.")
	smtpServer := flag.String("smtpserver", "", "The SMTP server to use for sending report emails.")
	smtpPort := flag.Int("smtpport", DefaultSMTPPort, "The port to use when connecting to the SMTP server.")
	smtpUsername := flag.String("smtpusername", "", "The username to use when connecting to the SMTP server.")
	smtpPassword := flag.String("smtppassword", "", "The password/secret to use when connecting to the SMTP server.")
	smtpAuthMethod := flag.String("smtpauthmethod", "", "The Auth method used by the SMTP server: plain or crammd5. No authentication is used by default.")
	mailTo := flag.String("mailto", "", "The email address to send reports to, comma delimited.")
	mailFrom := flag.String("mailfrom", "", "The email address reports are send from.")

	// Define the Usage function, which prints to Stderr
	// helpful information about the tool.
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
	if *domain == "" {
		log.Fatal("FATAL: An Alma API Server domain is required. https://developers.exlibrisgroup.com/alma/apis/#calling")
	}
	if *key == "" {
		log.Fatal("FATAL: An Alma API Key is required. https://developers.exlibrisgroup.com/alma/apis/#defining")
	}
	if *jobPath == "" {
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
			log.Fatal("FATAL: The email address reports are sent from must be provided if using the email option.")
		}
		if *smtpAuthMethod != "" {
			if *smtpAuthMethod != "crammd5" && *smtpAuthMethod != "plain" {
				log.Fatal("FATAL: 'crammd5' and 'plain' are the only supported auth methods for connecting to the SMTP server.")
			}
		}
	}

	// Create a buffer to store the email message.
	// The email report is a copy of the log messages.
	emailMessage := new(bytes.Buffer)

	// Split log output to both stderr and the email message.
	log.SetOutput(io.MultiWriter(os.Stderr, emailMessage))

	// Add the arguments to the output for later debugging.
	log.Println(*name)
	log.Println("Using alma-api-job-runner version", version)
	log.Println("Alma API server domain (domain):", *domain)
	log.Println("Job URL (url):", *jobPath)
	log.Println("Parameters file (params):", *params)
	log.Println("Sending email (email):", *sendEmail)

	// A closure to optionally send the report email and exit
	// with a non-zero error code.
	optionalEmailAndQuit := func() {
		if *sendEmail {
			err := SendEmail(*name+" -- error", emailMessage, *smtpServer, *smtpPort, *mailTo, *mailFrom, *smtpUsername, *smtpPassword, *smtpAuthMethod)
			if err != nil {
				log.Println(err)
			}
		}
		os.Exit(1)
	}

	// Build the request to the Alma API.
	jobURL, err := url.Parse(fmt.Sprintf("https://%v%v", *domain, *jobPath))
	if err != nil {
		log.Println("Error building final url from arguments: ", err)
		optionalEmailAndQuit()
	}
	log.Println("Going to submit parameters to:", jobURL)

	// Load the parameters XML file.
	// This is done to check that the XML is well formed and valid.
	loadedParams, err := LoadParameters(*params)
	if err != nil {
		log.Println("Error loading parameters: ", err)
		optionalEmailAndQuit()
	}

	// Log the parameters.
	log.Println("Parameters:")
	for _, param := range loadedParams.Parameters {
		log.Printf(" %v: %v\n", param.Name.Value, param.Value)
	}

	// Retry for max retries.
	jobInstanceLink, err := RetrySubmitJob(*maxRetries, jobURL, *timeout, *key, loadedParams)
	if err != nil {
		log.Println("Error when submitting job: ", err)
		optionalEmailAndQuit()
	}
	log.Println("Successful job submission.")

	instanceURL, err := url.Parse(jobInstanceLink)
	if err != nil {
		log.Printf("Error parsing instance url (%v) from job additional info: %v\n", jobInstanceLink, err)
		optionalEmailAndQuit()
	}
	log.Println("Going to monitor job at: ", instanceURL)

	instance, err := MonitorJobInstance(instanceURL, *timeout, *key)
	if err != nil {
		log.Println("Error monitoring job instance: ", err)
		optionalEmailAndQuit()
	}

	// Print the XML output of the final job instance to stdout.
	marshaledInstance, err := xml.MarshalIndent(instance, "", "  ")
	if err == nil {
		fmt.Println(string(marshaledInstance))
		_, err := emailMessage.Write(marshaledInstance)
		if err != nil {
			log.Println("Error writing XML representation to email message: ", err)
		}
	}

	if *sendEmail {
		subject := fmt.Sprintf("%v -- %v", *name, instance.Status.Desc)
		err := SendEmail(subject, emailMessage, *smtpServer, *smtpPort, *mailTo, *mailFrom, *smtpUsername, *smtpPassword, *smtpAuthMethod)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Email sent successfully")
		}
	}
}

// LoadParameters reads and unmarshals the contents of the params file.
func LoadParameters(params string) (loadedParams AlmaJob, err error) {
	// Get the absolute path of params, not strictly necessary
	// but it makes error messages more clear.
	paramsFilePath, err := filepath.Abs(params)
	if err != nil {
		return loadedParams, err
	}
	// Open the file for reading.
	paramsFile, err := os.Open(paramsFilePath)
	if err != nil {
		return loadedParams, err
	}
	// Defer the file close to the end of the function.
	defer paramsFile.Close()
	// Decode the file into an AlmaJob struct.
	decoder := xml.NewDecoder(paramsFile)
	err = decoder.Decode(&loadedParams)
	if err != nil {
		return loadedParams, err
	}
	return loadedParams, nil
}

// RetrySubmitJob retries SubmitJob max retries times.
func RetrySubmitJob(maxRetries int, url *url.URL, timeout int, key string, params AlmaJob) (jobInstanceLink string, err error) {
	for retry := 0; retry < maxRetries; retry++ {
		// Submit the Job, get the job instance ID back.
		jobInstanceLink, err := SubmitJob(url, timeout, key, params)
		if err != nil {
			// We encountered some error, retry with backoff.
			log.Println("Failed to submit job: ", err)
			sleepDur := time.Duration((retry+1)*(retry+1)) * time.Second
			log.Printf("Retrying in %v (%v/%v)\n", sleepDur, retry+1, maxRetries)
			time.Sleep(sleepDur)
			continue
		}
		return jobInstanceLink, nil
	}
	return "", fmt.Errorf("maximum number of retries reached")
}

// SubmitJob sends a POST HTTP request to the Alma API to execute the job.
func SubmitJob(url *url.URL, timeout int, key string, params AlmaJob) (jobInstanceLink string, err error) {
	// Setup the job parameter data as a io.Reader.
	marshaledParams := new(bytes.Buffer)
	encoder := xml.NewEncoder(marshaledParams)
	err = encoder.Encode(params)
	if err != nil {
		return "", err
	}

	// Setup an HTTP client with a timeout.
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Setup the request.
	request, err := http.NewRequest("POST", url.String(), marshaledParams)
	if err != nil {
		return "", err
	}
	request.Header.Add("Authorization", "apikey "+key)
	request.Header.Add("Content-Type", "application/xml")

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

	// If the response was a 400 error, we can (usually) parse the returned XML.
	if resp.StatusCode == 400 {
		apiErr := &APIError{}
		decoder := xml.NewDecoder(resp.Body)
		err := decoder.Decode(apiErr)
		// If the decode failed, still drain and close the response body.
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("alma API request failed, HTTP status %v, couldn't read body: %v", resp.Status, err)
		}
		return "", fmt.Errorf("alma API request failed, HTTP status %v, %v", resp.Status, apiErr.Collapse())
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

	// Decode the job and return the job instance ID.
	returnedJob := &AlmaJob{}
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(returnedJob)
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	return returnedJob.AdditionalInfo.Link, nil
}

// MonitorJobInstance will request the job instance until the job is complete or
// approximately 23 hours passes.
func MonitorJobInstance(url *url.URL, timeout int, key string) (instance *AlmaJobInstance, err error) {
	for i := 1; i < 2761; i++ {
		instance, err := GetJobInstance(url, timeout, key)
		if err != nil {
			return instance, err
		}
		log.Println("Job Status: ", instance.Status.Desc)
		if instance.EndTime != "" && instance.Status.Value != "FINALIZING" {
			return instance, nil
		}
		time.Sleep(30 * time.Second)
	}
	return instance, fmt.Errorf("job monitor has been running for 23 hours, exiting")
}

// GetJobInstance sends a GET HTTP request to the Alma API to get job instance data.
func GetJobInstance(url *url.URL, timeout int, key string) (instance *AlmaJobInstance, err error) {
	instance = &AlmaJobInstance{}

	// Setup an HTTP client with a timeout.
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Setup the request.
	request, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return instance, err
	}
	request.Header.Add("Authorization", "apikey "+key)

	// Do the request.
	// On error, drain and close the response body.
	resp, err := client.Do(request)
	if err != nil {
		if resp != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
		return instance, err
	}

	// Log the remaning number of API calls.
	remainingCalls := resp.Header.Get("X-Exl-Api-Remaining")
	if remainingCalls != "" {
		log.Printf("%v Alma API calls remaining.\n", remainingCalls)
	}

	// If the response was a 400 error, we can (usually) parse the returned XML.
	if resp.StatusCode == 400 {
		apiErr := &APIError{}
		decoder := xml.NewDecoder(resp.Body)
		err := decoder.Decode(apiErr)
		// If the decode failed, still drain and close the response body.
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
		if err != nil {
			return instance, fmt.Errorf("alma API request failed, HTTP status %v, couldn't read body: %v", resp.Status, err)
		}
		return instance, fmt.Errorf("alma API request failed, HTTP status %v, %v", resp.Status, apiErr.Collapse())
	}

	// If the Status != OK, there was an error we didn't catch yet.
	if resp.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return instance, fmt.Errorf("alma API request failed: %v couldn't read body: %v", resp.Status, err)
		}
		return instance, fmt.Errorf("alma API request failed: %v - %v", resp.Status, string(bodyBytes))
	}

	// Decode the job and return the job instance ID.
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(instance)
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	if err != nil {
		return instance, err
	}
	return instance, nil
}

// SendEmail sends an email using the provided configuration.
func SendEmail(subject string, emailMessage *bytes.Buffer, smtpServer string, smtpPort int, mailTo, mailFrom, smtpUsername, smtpPassword, smtpAuthMethod string) error {
	var auth smtp.Auth
	if smtpAuthMethod == "crammd5" {
		auth = smtp.CRAMMD5Auth(smtpUsername, smtpPassword)
	} else if smtpAuthMethod == "plain" {
		auth = smtp.PlainAuth("", smtpUsername, smtpPassword, smtpServer)
	}
	to := TrimSpaceAll(strings.Split(mailTo, ","))
	finalMsg := new(bytes.Buffer)
	finalMsg.WriteString(fmt.Sprintf("To: %v\r\n", strings.Join(to, ", ")))
	finalMsg.WriteString(fmt.Sprintf("Subject: %v\r\n", subject))
	finalMsg.WriteString("\r\n")
	bodyMsg := bytes.ReplaceAll(emailMessage.Bytes(), []byte("\n"), []byte("\r\n"))
	finalMsg.Write(bodyMsg)
	return smtp.SendMail(smtpServer+":"+strconv.Itoa(smtpPort), auth, mailFrom, to, finalMsg.Bytes())
}

// TrimSpaceAll returns a version of trimMe where each element has been TrimSpace'd.
func TrimSpaceAll(trimMe []string) []string {
	var trimmed []string
	for _, cur := range trimMe {
		trimmed = append(trimmed, strings.TrimSpace(cur))
	}
	return trimmed
}
