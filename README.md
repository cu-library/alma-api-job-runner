# alma-api-job-runner

[![Go Report Card](https://goreportcard.com/badge/github.com/cu-library/alma-api-job-runner)](https://goreportcard.com/report/github.com/cu-library/alma-api-job-runner)

This CLI tool simplifies submitting jobs to the Alma Jobs API.

It was designed so that system administrators could easily set up a cron job to schedule jobs in Alma which cannot be scheduled normally.

Set up job using the directions from this blog post: https://developers.exlibrisgroup.com/blog/Working-with-the-Alma-Jobs-API/.
Save the parameters in the "API Information" -> "XML" panel as an XML file.
Then run the tool, providing the:
* domain name of your Alma API server (https://developers.exlibrisgroup.com/alma/apis/#calling)
* Alma API Key (https://developers.exlibrisgroup.com/manage/keys/, make sure the API Key is configured with the "Configuration - Production Read/write" permission.
* URL (without the leading 'POST ') from the "API Information" -> "URL" panel.
* The path to the parameters file.

This tool:
* supports sending an email report using SMTP.
* can pull parameters from the environment using environment variables.
* automatically retries requests that fail, with an exponentially increasing backoff.

## Feedback welcome!

If you want an additional feature or find a bug, please add new issues here: https://github.com/cu-library/alma-api-job-runner/issues.


```
alma-api-job-runner:
Run a manual job in Alma using the Jobs API.

  -domain string
        The domain of the Alma API server URL to use. Required. (ex: api-ca.hosted.exlibrisgroup.com)
  -email
        Send an email report.
  -key string
        The Alma API key. Required.
  -mailfrom string
        The email address reports are send from.
  -mailto string
        The email address to send reports to, comma delimited.
  -name string
        The name for the job, used for logging and reports only. (default "Alma API Job Runner")
  -params string
        A file storing the XML representation of the job's parameters. Required.
  -retries int
        If calling the Alma API results in an error, how many times will the job be resubmitted. (default 5)
  -smtpauthmethod string
        The Auth method used by the SMTP server: plain or crammd5. No authentication is used by default.
  -smtppassword string
        The password/secret to use when connecting to the SMTP server.
  -smtpport int
        The port to use when connecting to the SMTP server. (default 25)
  -smtpserver string
        The SMTP server to use for sending report emails.
  -smtpusername string
        The username to use when connecting to the SMTP server.
  -timeout int
        The number of seconds to wait on the Alma API when submitting requests. (default 10)
  -url string
        The URL to which the job's parameters should be POST'd. Starts with a /. Required.
  Environment variables read when flag is unset:
  ALMA_API_JOB_RUNNER_DOMAIN
  ALMA_API_JOB_RUNNER_EMAIL
  ALMA_API_JOB_RUNNER_KEY
  ALMA_API_JOB_RUNNER_MAILFROM
  ALMA_API_JOB_RUNNER_MAILTO
  ALMA_API_JOB_RUNNER_NAME
  ALMA_API_JOB_RUNNER_PARAMS
  ALMA_API_JOB_RUNNER_RETRIES
  ALMA_API_JOB_RUNNER_SMTPAUTHMETHOD
  ALMA_API_JOB_RUNNER_SMTPPASSWORD
  ALMA_API_JOB_RUNNER_SMTPPORT
  ALMA_API_JOB_RUNNER_SMTPSERVER
  ALMA_API_JOB_RUNNER_SMTPUSERNAME
  ALMA_API_JOB_RUNNER_TIMEOUT
  ALMA_API_JOB_RUNNER_URL
```
