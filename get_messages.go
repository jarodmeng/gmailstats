package gmailstats

import (
	"log"
	"runtime"
)

// Create a work queue consisting of MessageWork using the MessageIds
// field of the calling GmailStats instance.
func (gs *GmailStats) createMessageWorkQueue() chan *MessageWork {
	if len(gs.MessageIds) == 0 {
		log.Fatalln("MessageIds field cannot be empty.")
	}

	messageWorkQueue := make(chan *MessageWork, len(gs.MessageIds))
	for _, mid := range gs.MessageIds {
		messageWork := &MessageWork{
			Id: mid.MessageId,
		}
		messageWorkQueue <- messageWork
	}
	close(messageWorkQueue)
	return messageWorkQueue
}

func (gs *GmailStats) GetMessages(verbose bool) *GmailStats {
	nMessageWorkers := runtime.NumCPU()
	messageOutput := make(chan *Message)

	messageWorkQueue := gs.createMessageWorkQueue()

	messageWorkerManager := NewMessageWorkerManager(gs, nMessageWorkers, messageWorkQueue, messageOutput)
	messageWorkerManager.Start(verbose)

	gs.Messages = make([]*Message, 0)
	for m := range messageOutput {
		gs.Messages = append(gs.Messages, m)
	}

	return gs
}
