package gmailstats

import (
	"time"

	gmail "google.golang.org/api/gmail/v1"
)

const (
	defaultClientSecret   = `{"installed":{"client_id":"956558626434-fsr1flbgemibocf40nqotjggalbad08u.apps.googleusercontent.com","project_id":"gmailstats-141502","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://accounts.google.com/o/oauth2/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_secret":"f6lKpgovIdd_RhvVPFDl8ifJ","redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}`
	defaultGmailTokenFile = "gmailstats_token.json"
	defaultGmailScope     = gmail.GmailReadonlyScope
	defaultGmailUser      = "me"
)

type MessageId struct {
	MessageId string `json:"messageid"`
	ThreadId  string `json:"threadid",omitempty`
}

type MessageHeader struct {
	FromEmail   string   `json:"fromemail",omitempty`
	ToEmails    []string `json:"toemails",omitempty`
	CcEmails    []string `json:"ccemails,omitempty"`
	BccEmails   []string `json:"bccemails,omitempty"`
	MailingList string   `json:"mailinglist",omitempty`
}

type MessageText struct {
	Subject  string `json:"subject",omitempty`
	Snippet  string `json:"snippet",omitempty`
	BodyText string `json:"bodytext",omitempty`
}

type JSONTime time.Time

type MessageTime struct {
	Time JSONTime `json:"time",omitempty`
}

type Message struct {
	Id     MessageId     `json:"id"`
	Time   MessageTime   `json:"time",omitempty`
	Header MessageHeader `json:"header",omitempty`
	Text   MessageText   `json:"text",omitempty`
}

type GmailStats struct {
	service    *gmail.Service
	MessageIds []*MessageId
	Messages   []*Message
}
