package gmailstats

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"

	gmail "google.golang.org/api/gmail/v1"
)

func matchEmail(s string) string {
	const emailRawRegex = `(?:^|(?:^.+ <)|^<)([[:alnum:]-+@.]+)(?:$|(?:>$))`
	emailRegex, _ := regexp.Compile(emailRawRegex)
	email := emailRegex.FindStringSubmatch(s)
	if len(email) < 2 {
		return s
	}
	return email[1]
}

func matchAllEmails(s string) []string {
	out := make([]string, 0)
	emailStrings := strings.Split(s, ", ")
	for _, e := range emailStrings {
		email := matchEmail(e)
		out = append(out, email)
	}
	return out
}

func matchMailingList(s string) string {
	const mlRawRegex = `^list ([[:alnum:]-@.]+);.+`
	mlRegex := regexp.MustCompile(mlRawRegex)
	ml := mlRegex.FindStringSubmatch(s)
	if len(ml) < 2 {
		return s
	}
	return ml[1]
}

func GetMessage(service *gmail.Service, Id string) *Message {
	fmt.Printf("Processing message id %s.\n", Id)
	mr, _ := service.Users.Messages.Get(defaultGmailUser, Id).Do()

	messageId := MessageId{
		MessageId: mr.Id,
		ThreadId:  mr.ThreadId,
	}

	messageTime := MessageTime{
		Time: mr.InternalDate / 1000,
	}

	messageHeader := MessageHeader{}
	messageText := MessageText{
		Snippet:  mr.Snippet,
		BodyText: "test",
	}
	for _, h := range mr.Payload.Headers {
		switch h.Name {
		case "From":
			messageHeader.FromEmail = matchEmail(h.Value)
		case "To":
			messageHeader.ToEmails = matchAllEmails(h.Value)
		case "Cc":
			messageHeader.CcEmails = matchAllEmails(h.Value)
		case "Bcc":
			messageHeader.BccEmails = matchAllEmails(h.Value)
		case "Mailing-list":
			messageHeader.MailingList = matchMailingList(h.Value)
		case "Subject":
			messageText.Subject = h.Value
		}
	}

	message := &Message{
		Id:     messageId,
		Time:   messageTime,
		Header: messageHeader,
		Text:   messageText,
	}

	return message
}

func (gs *GmailStats) createMessageWorkQueue() chan MessageWorkRequest {
	messageWorkQueue := make(chan MessageWorkRequest, len(gs.MessageIds))
	for _, mid := range gs.MessageIds {
		messageWork := MessageWorkRequest{
			Id: mid.MessageId,
		}
		messageWorkQueue <- messageWork
	}
	close(messageWorkQueue)
	return messageWorkQueue
}

func (gs *GmailStats) GetMessages() *GmailStats {
	nMessageWorkers := runtime.NumCPU()
	messageWorkQueue := gs.createMessageWorkQueue()
	messageOutput := getMessages(gs, nMessageWorkers, messageWorkQueue)
	gs.Messages = make([]*Message, 0)
	for m := range messageOutput {
		gs.Messages = append(gs.Messages, m)
	}
	return gs
}

func getMessages(gs *GmailStats, nMessageWorkers int, messageWorkQueue chan MessageWorkRequest) chan *Message {
	team := make(chan *MessageWorker, nMessageWorkers)
	messageOutput := make(chan *Message)

	var wg sync.WaitGroup

	for i := 0; i < nMessageWorkers; i++ {
		wg.Add(1)
		messageWorker := NewMessageWorker(gs, team, messageOutput, &wg)
		messageWorker.Start()
	}

	go func() {
		for work := range messageWorkQueue {
			messageWorker := <-team
			messageWorker.messageWorkQueue <- work
		}
		for messageWorker := range team {
			close(messageWorker.messageWorkQueue)
		}
		close(messageOutput)
	}()

	go func() {
		wg.Wait()
		close(team)
	}()

	return messageOutput
}

type MessageWorker struct {
	gs               *GmailStats
	team             chan *MessageWorker
	messageOutput    chan *Message
	waitGroup        *sync.WaitGroup
	messageWorkQueue chan MessageWorkRequest
}

func NewMessageWorker(gs *GmailStats, team chan *MessageWorker, messageOutput chan *Message, wg *sync.WaitGroup) *MessageWorker {
	messageWorker := &MessageWorker{
		gs:               gs,
		team:             team,
		messageOutput:    messageOutput,
		waitGroup:        wg,
		messageWorkQueue: make(chan MessageWorkRequest),
	}

	return messageWorker
}

func (messageWorker *MessageWorker) Start() {
	go func() {
		messageWorker.team <- messageWorker
		for work := range messageWorker.messageWorkQueue {
			message := GetMessage(messageWorker.gs.service, work.Id)
			messageWorker.messageOutput <- message
			messageWorker.team <- messageWorker
		}
		messageWorker.waitGroup.Done()
	}()
}
