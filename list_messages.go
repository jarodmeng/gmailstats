package gmailstats

import (
	"log"
	"math"

	gmail "google.golang.org/api/gmail/v1"
)

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
// MessageIds field of the calling GmailStats instance.
func (l *ListMessagesCall) Do() (*GmailStats, error) {
	const messageBatchSize = 500
	messages := make([]*MessageId, 0)
	resultsCountDown := l.maxResults

	minResults := func(rr *int64, limit int64) int64 {
		return int64(math.Min(float64(*rr), float64(limit)))
	}

	extractMessageToken := func(r *gmail.ListMessagesResponse) ([]*MessageId, string) {
		return extractMessageId(r.Messages), r.NextPageToken
	}

	call := l.gs.service.Users.Messages.List(defaultGmailUser).MaxResults(minResults(&resultsCountDown, messageBatchSize)).Q(l.q)
	r0, err := call.Do()
	if err != nil {
		return l.gs, err
	}
	ms, nextToken := extractMessageToken(r0)
	messages = append(messages, ms...)
	resultsCountDown -= int64(len(ms))

	// Continue to get more results when the number of retrieved messages is less
	// than maxResults and there's still more to get.
	for int64(len(messages)) < l.maxResults && nextToken != "" {
		r1, err := call.PageToken(nextToken).Do()
		if err != nil {
			return l.gs, err
		}
		ms, nextToken = extractMessageToken(r1)
		messages = append(messages, ms...)
		resultsCountDown -= int64(len(ms))
	}

	l.gs.MessageIds = messages

	return l.gs, nil
}
