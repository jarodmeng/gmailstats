package gmailstats

import (
	"fmt"
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

func (gs *GmailStats) GetMessages(write bool) *GmailStats {
	nMessageWorkers := runtime.NumCPU()
	// output channel
	messageOutput := make(chan *Message)
	// input channel
	messageWorkQueue := gs.createMessageWorkQueue()

	// Start a manager that takes in work input, organizes a worker pool and sends
	// output to the output channel
	messageWorkerManager := gs.newMessageWorkerManager(nMessageWorkers, messageWorkQueue, messageOutput)
	messageWorkerManager.start()

	// Move Message from output channel to the Messages field in GmailStats
	gs.Messages = make([]*Message, 0)
	if write {
		if gs.MessagesFile == nil {
			fmt.Println("No MessagesFile found. Use messages.json by default.")
			gs.OpenMessagesFile("messages.json")
		}
		defer gs.MessagesFile.Close()
	}
	for m := range messageOutput {
		gs.Messages = append(gs.Messages, m)
		if write {
			if err := m.writeJSONToFile(gs.MessagesFile); err != nil {
				log.Fatalf("Unable to write JSON file: %v.\n", err)
			}
		}
	}

	return gs
}
