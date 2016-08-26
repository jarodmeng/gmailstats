// Package gmailstats offers a collection of facilities to interact with Google
// Gmail API.
package gmailstats

import (
	"log"

	"github.com/jarodmeng/googleauth"
	gmail "google.golang.org/api/gmail/v1"
)

func createGmailService(tokenFile string, scope string) (*gmail.Service, error) {
	b := []byte(defaultClientSecret)

	gmailClient, err := googleauth.CreateClient(b, tokenFile, scope)
	if err != nil {
		return nil, err
	}

	gmailService, err := gmail.New(gmailClient)
	if err != nil {
		return nil, err
	}

	return gmailService, nil
}

// New creates a GmailStats instance with a service object ready to make API
// calls.
func New() *GmailStats {
	gmailService, err := createGmailService(defaultGmailTokenFile, defaultGmailScope)
	if err != nil {
		log.Fatalf("Unable to create Gmail service: %v.\n", err)
	}

	gs := &GmailStats{
		service: gmailService,
	}

	return gs
}
