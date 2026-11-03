package grants

import "net/http"

// ServeToken serves a token request outside the /token route, for DIMO's web3
// login, which exchanges a signed challenge in a single call.
func (h *Handler) ServeToken(w http.ResponseWriter, r *http.Request) {
	h.handleToken(w, r)
}
