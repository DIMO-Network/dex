package server

import "net/http"

func (t *templates) web3login(r *http.Request, w http.ResponseWriter, challengeURL string, verifyURL string, authID string, infuraID string) error {
	data := struct {
		ChallengeURL string
		VerifyURL    string
		AuthID       string
		ReqPath      string
		InfuraID     string
	}{challengeURL, verifyURL, authID, r.URL.Path, infuraID}
	return renderTemplate(w, t.web3Tmpl, data)
}
