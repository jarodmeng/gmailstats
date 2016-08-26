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

func (gs *GmailStats) ListMessages() *ListMessagesCall {
	lmc := &ListMessagesCall{
		gs:         gs,
		maxResults: 100,
		q:          "-is:chat",
	}

	return lmc
}

func (lmc *ListMessagesCall) MaxResults(maxResults int64) *ListMessagesCall {
	if maxResults <= 0 {
		log.Fatalln("maxResults must be a positive integer.")
	}

	lmc.maxResults = maxResults
	return lmc
}

func (lmc *ListMessagesCall) Q(q string) *ListMessagesCall {
	lmc.q = lmc.q + " " + q
	return lmc
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

func (lmc *ListMessagesCall) Do() (*GmailStats, error) {
	const numMessages = 500
	messages := make([]*MessageId, 0)
	remainingResults := lmc.maxResults

	minResults := func(rr *int64, limit int64) int64 {
		return int64(math.Min(float64(*rr), float64(limit)))
	}

	call := lmc.gs.service.Users.Messages.List(defaultGmailUser).MaxResults(minResults(&remainingResults, numMessages)).Q(lmc.q)
	r0, err := call.Do()
	if err != nil {
		return lmc.gs, err
	}
	ms := extractMessageId(r0.Messages)
	nextToken := r0.NextPageToken
	messages = append(messages, ms...)
	remainingResults = remainingResults - int64(len(ms))

	for int64(len(messages)) < lmc.maxResults && nextToken != "" {
		r1, err := call.PageToken(nextToken).Do()
		if err != nil {
			return lmc.gs, err
		}
		ms = extractMessageId(r1.Messages)
		nextToken = r1.NextPageToken
		messages = append(messages, ms...)
		remainingResults = remainingResults - int64(len(ms))
	}

	lmc.gs.MessageIds = messages

	return lmc.gs, nil
}
