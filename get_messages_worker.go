package gmailstats

import (
	"log"
	"sync"
)

type messageWork struct {
	Id string
}

type messageWorker struct {
	gs        *GmailStats
	team      chan *messageWorker
	output    chan *Message
	finish    *sync.WaitGroup
	workQueue chan *messageWork
}

func newMessageWorker(gs *GmailStats, team chan *messageWorker, out chan *Message, finish *sync.WaitGroup) *messageWorker {
	mw := &messageWorker{
		gs:        gs,
		team:      team,
		output:    out,
		finish:    finish,
		workQueue: make(chan *messageWork),
	}

	return mw
}

func (mw *messageWorker) start(verbose bool) {
	go func() {
		mw.team <- mw
		for work := range mw.workQueue {
			mw.processMessage(work, verbose)
			mw.team <- mw
		}
		mw.finish.Done()
	}()
}

// processMessage gets detailed information about a message and packages them into
// a Message object
func (mw *messageWorker) processMessage(messageWork *messageWork, verbose bool) {
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

type messageWorkerManager struct {
	gs        *GmailStats
	nWorkers  int
	team      chan *messageWorker
	finish    *sync.WaitGroup
	workQueue chan *messageWork
	output    chan *Message
}

func newMessageWorkerManager(gs *GmailStats, n int, wq chan *messageWork, out chan *Message) *messageWorkerManager {
	messageWorkerManager := &messageWorkerManager{
		gs:        gs,
		nWorkers:  n,
		team:      make(chan *messageWorker, n),
		finish:    &sync.WaitGroup{},
		workQueue: wq,
		output:    out,
	}

	return messageWorkerManager
}

func (mwm *messageWorkerManager) start(verbose bool) {
	for i := 0; i < mwm.nWorkers; i++ {
		mwm.finish.Add(1)
		messageWorker := newMessageWorker(mwm.gs, mwm.team, mwm.output, mwm.finish)
		messageWorker.start(verbose)
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
