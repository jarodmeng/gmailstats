package gmailstats

import (
	"os"

	gmail "google.golang.org/api/gmail/v1"
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

type MessageTime struct {
	Time int64 `json:"time",omitempty`
}

type Message struct {
	Id     *MessageId     `json:"id"`
	Time   *MessageTime   `json:"time",omitempty`
	Header *MessageHeader `json:"header",omitempty`
	Text   *MessageText   `json:"text",omitempty`
}

type GmailStats struct {
	service      *gmail.Service
	MessageIds   []*MessageId
	Messages     []*Message
	MessagesFile *os.File
}
