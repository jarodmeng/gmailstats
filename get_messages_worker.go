package gmailstats

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type messageWork struct {
	Id string
}

type messageWorker struct {
	mwm       *messageWorkerManager
	id        int
	workQueue chan *messageWork
}

func (mwm *messageWorkerManager) newMessageWorker(id int) *messageWorker {
	mw := &messageWorker{
		mwm:       mwm,
		id:        id,
		workQueue: make(chan *messageWork),
	}

	return mw
}

func (mw *messageWorker) start() {
	go func() {
		mw.mwm.team <- mw // add worker itself to team
		for work := range mw.workQueue {
			err := mw.processMessage(work) // process received work
			if err != nil {
				fmt.Printf("Error when processing message id %s: %v.\n", work.Id, err)
				fmt.Printf("Worker %d is now sleeping for 10 seconds.\n", mw.id)
				time.Sleep(10 * time.Second)
				mw2 := <-mw.mwm.team
				mw2.workQueue <- work
			}
			mw.mwm.team <- mw // add worker itself back to team
		}
		// When work queue is completed, sign off the worker
		fmt.Printf("messageWorker %d signs off.\n", mw.id)
		mw.mwm.finish.Done()
	}()
}

func (mw *messageWorker) processMessage(w *messageWork) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	mr, _ := mw.mwm.gs.service.Users.Messages.Get(defaultGmailUser, w.Id).Do()

	messageId := &MessageId{
		MessageId: mr.Id,
		ThreadId:  mr.ThreadId,
	}

	messageTime := &MessageTime{
		Time: mr.InternalDate / 1000,
	}

	messageHeader := &MessageHeader{}
	messageText := &MessageText{
		Snippet:  mr.Snippet,
		BodyText: "",
	}
	for _, h := range mr.Payload.Headers {
		switch strings.ToLower(h.Name) {
		case "from":
			messageHeader.FromEmail = matchEmail(h.Value)
		case "to":
			messageHeader.ToEmails = matchAllEmails(h.Value)
		case "cc":
			messageHeader.CcEmails = matchAllEmails(h.Value)
		case "bcc":
			messageHeader.BccEmails = matchAllEmails(h.Value)
		case "mailing-list":
			messageHeader.MailingList = matchMailingList(h.Value)
		case "subject":
			messageText.Subject = h.Value
		}
	}

	message := &Message{
		Id:     messageId,
		Time:   messageTime,
		Header: messageHeader,
		Text:   messageText,
	}

	fmt.Printf("Processed request of message id %s.\n", w.Id)
	mw.mwm.output <- message

	return nil
}

type messageWorkerManager struct {
	gs        *GmailStats         // parent GmailStats instance
	nWorkers  int                 // number of workers
	team      chan *messageWorker // a team organized as a channel of workers
	finish    *sync.WaitGroup     // worker registration and sign-off sheet
	workQueue chan *messageWork   // input as a channel of work
	output    chan *Message       // output as a channel of Message
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
		messageWorker := mwm.newMessageWorker(i) // create worker
		messageWorker.start()                    // start the newly-created worker
		mwm.finish.Add(1)                        // increment worker tracking
	}

	go func() {
		// This loop exhausts the input work.
		for work := range mwm.workQueue {
			messageWorker := <-mwm.team     // get a worker
			messageWorker.workQueue <- work // assign work to worker
		}
		// Once input work is done, close workers by closing workers's input queues.
		for messageWorker := range mwm.team {
			close(messageWorker.workQueue)
		}
		// Once all workers are closed, close output channel.
		close(mwm.output)
	}()

	go func() {
		mwm.finish.Wait() // wait for all workers to sign off
		close(mwm.team)
	}()
}
