package gmailstats

import (
	"log"
	"sync"
)

type MessageWork struct {
	Id string
}

type MessageWorker struct {
	gs        *GmailStats
	team      chan *MessageWorker
	output    chan *Message
	finish    *sync.WaitGroup
	workQueue chan *MessageWork
}

func NewMessageWorker(gs *GmailStats, team chan *MessageWorker, out chan *Message, finish *sync.WaitGroup) *MessageWorker {
	mw := &MessageWorker{
		gs:        gs,
		team:      team,
		output:    out,
		finish:    finish,
		workQueue: make(chan *MessageWork),
	}

	return mw
}

func (mw *MessageWorker) Start(verbose bool) {
	go func() {
		mw.team <- mw
		for work := range mw.workQueue {
			mw.ProcessMessage(work, verbose)
			mw.team <- mw
		}
		mw.finish.Done()
	}()
}

// ProcessMessage gets detailed information about a message and packages them into
// a Message object
func (mw *MessageWorker) ProcessMessage(messageWork *MessageWork, verbose bool) {
	mr, _ := mw.gs.service.Users.Messages.Get(defaultGmailUser, messageWork.Id).Do()

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

	if verbose {
		log.Printf("Processed request of message id %s.\n", messageWork.Id)
	}

	mw.output <- message
}

type MessageWorkerManager struct {
	gs        *GmailStats
	nWorkers  int
	team      chan *MessageWorker
	finish    *sync.WaitGroup
	workQueue chan *MessageWork
	output    chan *Message
}

func NewMessageWorkerManager(gs *GmailStats, n int, wq chan *MessageWork, out chan *Message) *MessageWorkerManager {
	messageWorkerManager := &MessageWorkerManager{
		gs:        gs,
		nWorkers:  n,
		team:      make(chan *MessageWorker, n),
		finish:    &sync.WaitGroup{},
		workQueue: wq,
		output:    out,
	}

	return messageWorkerManager
}

func (mwm *MessageWorkerManager) Start(verbose bool) {
	for i := 0; i < mwm.nWorkers; i++ {
		mwm.finish.Add(1)
		messageWorker := NewMessageWorker(mwm.gs, mwm.team, mwm.output, mwm.finish)
		messageWorker.Start(verbose)
	}

	go func() {
		for work := range mwm.workQueue {
			messageWorker := <-mwm.team
			messageWorker.workQueue <- work
		}
		for messageWorker := range mwm.team {
			close(messageWorker.workQueue)
		}
		close(mwm.output)
	}()

	go func() {
		mwm.finish.Wait()
		close(mwm.team)
	}()
}
