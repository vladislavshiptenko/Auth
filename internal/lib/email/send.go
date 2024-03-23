package email

import (
	"fmt"
	"net/http"
)

const (
	format = "json"
	listId = "1"
)

func FormSendEmail(apiKey string, senderName string, senderEmail string, recipientEmail string, subject string, body string) (*http.Request, error) {
	req, err := http.NewRequest("POST", "https://api.unisender.com/ru/api/sendEmail", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("api_key", apiKey)
	q.Add("sender_name", senderName)
	q.Add("sender_email", senderEmail)
	q.Add("email", fmt.Sprintf("User <%s>", recipientEmail))
	q.Add("subject", subject)
	q.Add("body", body)
	q.Add("format", format)
	q.Add("list_id", listId)
	req.URL.RawQuery = q.Encode()

	return req, nil
}
