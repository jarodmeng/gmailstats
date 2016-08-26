package gmailstats

import gmail "google.golang.org/api/gmail/v1"

const (
	defaultClientSecret   = `{"installed":{"client_id":"956558626434-fsr1flbgemibocf40nqotjggalbad08u.apps.googleusercontent.com","project_id":"gmailstats-141502","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://accounts.google.com/o/oauth2/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_secret":"f6lKpgovIdd_RhvVPFDl8ifJ","redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}`
	defaultGmailTokenFile = "gmailstats_token.json"
	defaultGmailScope     = gmail.GmailReadonlyScope
	defaultGmailUser      = "me"
)
