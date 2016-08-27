package gmailstats

import (
	"log"
	"runtime"
)

// Create a work queue consisting of messageWork using the MessageIds
// field of the calling GmailStats instance.
func (gs *GmailStats) createMessageWorkQueue() chan *messageWork {
	if len(gs.MessageIds) == 0 {
		log.Fatalln("MessageIds field cannot be empty.")
	}

	messageWorkQueue := make(chan *messageWork, len(gs.MessageIds))
	for _, mid := range gs.MessageIds {
		messageWork := &messageWork{
			Id: mid.MessageId,
		}
		messageWorkQueue <- messageWork
	}
	close(messageWorkQueue)
	return messageWorkQueue
}

func (gs *GmailStats) GetMessages() *GmailStats {
	nMessageWorkers := runtime.NumCPU()
	messageOutput := make(chan *Message)

	messageWorkQueue := gs.createMessageWorkQueue()

	messageWorkerManager := gs.newMessageWorkerManager(nMessageWorkers, messageWorkQueue, messageOutput)
	messageWorkerManager.start()

	gs.Messages = make([]*Message, 0)
	for m := range messageOutput {
		gs.Messages = append(gs.Messages, m)
	}

	return gs
}
