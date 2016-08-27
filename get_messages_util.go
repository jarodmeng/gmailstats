package gmailstats

import (
	"log"
	"regexp"
	"strings"
)

const (
	emailRawRegex       = `(?:^|(?:^.+ < *)|^<)([[:alnum:]-+_@.]+)(?:$|(?:>$))`
	mailingListRawRegex = `^list ([[:alnum:]-@.]+);.+`
)

func matchEmail(s string) string {
	emailRegex := regexp.MustCompile(emailRawRegex)
	email := emailRegex.FindStringSubmatch(s)
	if len(email) < 2 {
		log.Printf("Cannot parse this email: %s.\n", s)
		return s
	}
	return email[1]
}

func matchAllEmails(s string) []string {
	out := make([]string, 0)
	emailStrings := strings.Split(s, ", ")
	for _, e := range emailStrings {
		email := matchEmail(e)
		out = append(out, email)
	}
	return out
}

func matchMailingList(s string) string {
	mailingListRegex := regexp.MustCompile(mailingListRawRegex)
	mailingList := mailingListRegex.FindStringSubmatch(s)
	if len(mailingList) < 2 {
		log.Printf("Cannot parse this mailing list: %s.\n", s)
		return s
	}
	return mailingList[1]
}
