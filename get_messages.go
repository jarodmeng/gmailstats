package gmailstats

import (
	"fmt"
	"log"
)

type GetMessagesCall struct {
	gs     *GmailStats
	append bool
	write  bool
}

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

func (gs *GmailStats) GetMessages() *GetMessagesCall {
	g := &GetMessagesCall{
		gs:     gs,
		append: false,
		write:  false,
	}

	return g
}

func (g *GetMessagesCall) Append() *GetMessagesCall {
	g.append = true
	return g
}

func (g *GetMessagesCall) Write() *GetMessagesCall {
	g.write = true
	return g
}

func (g *GetMessagesCall) Do(numCPU int) *GmailStats {
	nMessageWorkers := numCPU
	// output channel
	messageOutput := make(chan *Message)
	// input channel
	messageWorkQueue := g.gs.createMessageWorkQueue()

	// Start a manager that takes in work input, organizes a worker pool and sends
	// output to the output channel
	messageWorkerManager := g.gs.newMessageWorkerManager(nMessageWorkers, messageWorkQueue, messageOutput)
	messageWorkerManager.start()

	// Move Message from output channel to the Messages field in GmailStats
	if !g.append {
		g.gs.Messages = make([]*Message, 0)
	}
	if g.write {
		if g.gs.MessagesFile == nil {
			fmt.Println("No MessagesFile found. Use messages.json by default.")
			g.gs.OpenMessagesFile("messages.json")
		}
		defer func() {
			g.gs.MessagesFile.Close()
			fmt.Println("MessagesFile closed.")
		}()
	}
	for m := range messageOutput {
		g.gs.Messages = append(g.gs.Messages, m)
		if g.write {
			if err := m.writeJSONToFile(g.gs.MessagesFile); err != nil {
				fmt.Printf("Unable to write JSON file: %v.\n", err)
			}
		}
	}

	return g.gs
}
