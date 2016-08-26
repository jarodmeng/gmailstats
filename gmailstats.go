// Package gmailstats offers a collection of facilities to interact with Google
// Gmail API.
package gmailstats

import (
	"log"
	"math"

	"github.com/jarodmeng/googleauth"
	gmail "google.golang.org/api/gmail/v1"
)

func createGmailService(tokenFile string, scope string) (*gmail.Service, error) {
	b := []byte(defaultClientSecret)

	gmailClient, err := googleauth.CreateClient(b, tokenFile, scope)
	if err != nil {
		return nil, err
	}

	gmailService, err := gmail.New(gmailClient)
	if err != nil {
		return nil, err
	}

	return gmailService, nil
}

// New creates a GmailStats instance with a service object ready to make API
// calls.
func New() *GmailStats {
	gmailService, err := createGmailService(defaultGmailTokenFile, defaultGmailScope)
	if err != nil {
		log.Fatalf("Unable to create Gmail service: %v.\n", err)
	}

	gs := &GmailStats{
		service: gmailService,
	}

	return gs
}

type ListMessagesCall struct {
	gs         *GmailStats
	q          string
	maxResults int64
}

// ListMessages lists a number of messages and store their message ids and
// thread ids in the calling GmailStats instance.
// The default number of messages retrieved is 100.
// Chat messages are excluded.
func (gs *GmailStats) ListMessages() *ListMessagesCall {
	l := &ListMessagesCall{
		gs:         gs,
		maxResults: 100,
		q:          "-is:chat",
	}

	return l
}

// MaxResults modifies the ListMessagesCall to retrieve a particular number of
// messages specified.
func (l *ListMessagesCall) MaxResults(maxResults int64) *ListMessagesCall {
	if maxResults <= 0 {
		log.Fatalln("maxResults must be a positive integer.")
	}

	l.maxResults = maxResults
	return l
}

// Q modifies the ListMessagesCall to only search for messages that match the
// provided query string.
func (l *ListMessagesCall) Q(q string) *ListMessagesCall {
	l.q = l.q + " " + q
	return l
}

func extractMessageId(ms []*gmail.Message) []*MessageId {
	messages := make([]*MessageId, 0)
	for _, m := range ms {
		mid := &MessageId{
			MessageId: m.Id,
			ThreadId:  m.ThreadId,
		}
		messages = append(messages, mid)
	}

	return messages
}

// Do executes the ListMessagesCall and stores the retrieved messages in the
// calling GmailStats instance.
func (l *ListMessagesCall) Do() (*GmailStats, error) {
	const numMessages = 500
	messages := make([]*MessageId, 0)
	remainingResults := l.maxResults

	minResults := func(rr *int64, limit int64) int64 {
		return int64(math.Min(float64(*rr), float64(limit)))
	}

	call := l.gs.service.Users.Messages.List(defaultGmailUser).MaxResults(minResults(&remainingResults, numMessages)).Q(l.q)
	r0, err := call.Do()
	if err != nil {
		return l.gs, err
	}
	ms := extractMessageId(r0.Messages)
	nextToken := r0.NextPageToken
	messages = append(messages, ms...)
	remainingResults = remainingResults - int64(len(ms))

	for int64(len(messages)) < l.maxResults && nextToken != "" {
		r1, err := call.PageToken(nextToken).Do()
		if err != nil {
			return l.gs, err
		}
		ms = extractMessageId(r1.Messages)
		nextToken = r1.NextPageToken
		messages = append(messages, ms...)
		remainingResults = remainingResults - int64(len(ms))
	}

	l.gs.MessageIds = messages

	return l.gs, nil
}
