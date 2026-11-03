package templates

import "net/http"

const tmplWeb3 = "web3.html"

func (t *Templates) Web3Login(r *http.Request, w http.ResponseWriter, challengeURL string, verifyURL string, authID string) error {
	data := struct {
		ChallengeURL string
		VerifyURL    string
		AuthID       string
		ReqPath      string
	}{challengeURL, verifyURL, authID, r.URL.Path}
	return renderTemplate(w, t.web3Tmpl, data)
}
