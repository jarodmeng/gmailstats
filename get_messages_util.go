package gmailstats

import (
	"log"
	"regexp"
	"strings"
)

func matchEmail(s string) string {
	const emailRawRegex = `(?:^|(?:^.+ < *)|^<)([[:alnum:]-+_@.]+)(?:$|(?:>$))`
	emailRegex, _ := regexp.Compile(emailRawRegex)
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
	const mlRawRegex = `^list ([[:alnum:]-@.]+);.+`
	mlRegex := regexp.MustCompile(mlRawRegex)
	ml := mlRegex.FindStringSubmatch(s)
	if len(ml) < 2 {
		log.Printf("Cannot parse this mailing list: %s.\n", s)
		return s
	}
	return ml[1]
}
