package gmailstats

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	emailSplitRawRegex  = `(, *|;)`
	emailRawRegex       = `(?:^|(?:^.+ *< *)|^<)([[:alnum:]-+_@.=%]+)(?:$|(?:>.*$))`
	mailingListRawRegex = `^list ([[:alnum:]-+_@.]+);.+`
)

func matchEmail(s string) string {
	emailRegex := regexp.MustCompile(emailRawRegex)
	email := emailRegex.FindStringSubmatch(s)
	if len(email) < 2 {
		fmt.Printf("Cannot parse this email: %s.\n", s)
		return s
	}
	return strings.ToLower(email[1])
}

func matchAllEmails(s string) []string {
	out := make([]string, 0)
	emailStrings := regexp.MustCompile(emailSplitRawRegex).Split(s, -1)
	for _, e := range emailStrings {
		if !strings.Contains(e, "@") {
			continue
		}
		email := matchEmail(e)
		out = append(out, email)
	}
	return out
}

func matchMailingList(s string) string {
	mailingListRegex := regexp.MustCompile(mailingListRawRegex)
	mailingList := mailingListRegex.FindStringSubmatch(s)
	if len(mailingList) < 2 {
		fmt.Printf("Cannot parse this mailing list: %s.\n", s)
		return s
	}
	return strings.ToLower(mailingList[1])
}

func (m *Message) writeJSONToFile(f *os.File) error {
	err := json.NewEncoder(f).Encode(m)
	if err != nil {
		return err
	}
	return nil
}
