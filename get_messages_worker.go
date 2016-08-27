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

func (mwm *messageWorkerManager) newMessageWorker() *messageWorker {
	mw := &messageWorker{
		gs:        mwm.gs,
		team:      mwm.team,
		output:    mwm.output,
		finish:    mwm.finish,
		workQueue: make(chan *messageWork),
	}

	return mw
}

func (mw *messageWorker) start() {
	go func() {
		mw.team <- mw
		for work := range mw.workQueue {
			mw.processMessage(work)
			mw.team <- mw
		}
		mw.finish.Done()
	}()
}

func (mw *messageWorker) processMessage(messageWork *messageWork) {
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

	log.Printf("Processed request of message id %s.\n", messageWork.Id)
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

func (gs *GmailStats) newMessageWorkerManager(n int, wq chan *messageWork, out chan *Message) *messageWorkerManager {
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

func (mwm *messageWorkerManager) start() {
	for i := 0; i < mwm.nWorkers; i++ {
		mwm.finish.Add(1)
		messageWorker := mwm.newMessageWorker()
		messageWorker.start()
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
