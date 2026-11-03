package server

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleInvalidFormPOSTCallbacks(t *testing.T) {
	httpServer, server := newTestServer(t, func(c *Config) {
		c.Storage = &emptyStorage{c.Storage}
	})
	defer httpServer.Close()

	const (
		noState      = "User session error."
		unknownState = "Requested resource does not exist."
	)
	tests := []struct {
		name string
		form url.Values
		want string
	}{
		{name: "no state", form: url.Values{}, want: noState},
		{name: "unknown state", form: url.Values{"state": {"abcd"}}, want: unknownState},
		{name: "unknown RelayState", form: url.Values{"RelayState": {"abcd"}}, want: unknownState},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/callback", strings.NewReader(tc.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			server.ServeHTTP(rr, req)

			require.Equal(t, http.StatusBadRequest, rr.Code, rr.Body.String())
			require.Contains(t, rr.Body.String(), tc.want)
		})
	}
}
