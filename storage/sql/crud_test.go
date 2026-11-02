//go:build cgo
// +build cgo

package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dexidp/dex/storage"
)

func TestDecoder(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`create table foo ( id integer primary key, bar blob );`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`insert into foo ( id, bar ) values (1, ?);`, []byte(`["a", "b"]`)); err != nil {
		t.Fatal(err)
	}
	var got []string
	if err := db.QueryRow(`select bar from foo where id = 1;`).Scan(decoder(&got)); err != nil {
		t.Fatal(err)
	}
	want := []string{"a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wanted %q got %q", want, got)
	}
}

func TestEncoder(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`create table foo ( id integer primary key, bar blob );`); err != nil {
		t.Fatal(err)
	}
	put := []string{"a", "b"}
	if _, err := db.Exec(`insert into foo ( id, bar ) values (1, ?)`, encoder(put)); err != nil {
		t.Fatal(err)
	}

	var got []byte
	if err := db.QueryRow(`select bar from foo where id = 1;`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	want := []byte(`["a","b"]`)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wanted %q got %q", want, got)
	}
}

// A database failure is not a missing client: it must neither fall through to the
// identity-api lookup nor come back as an empty client with no error.
func TestGetClientReturnsDatabaseErrors(t *testing.T) {
	c, err := (&SQLite3{File: ":memory:"}).open(slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}

	client, err := c.GetClient(context.Background(), "client")
	assert.Error(t, err)
	assert.Equal(t, storage.Client{}, client)
}

func TestGetDeveloperClient(t *testing.T) {
	tests := []struct {
		name           string
		clientID       string
		mockResponse   interface{}
		mockStatusCode int
		expectedClient storage.Client
		expectedError  string
	}{
		{
			name:     "Success",
			clientID: "0x1549487A47332da96cc3EE18D860aaE7b45E2eF3",
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"developerLicense": map[string]interface{}{
						"redirectURIs": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{"uri": "https://example.com/callback"},
							},
						},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			expectedClient: storage.Client{
				ID:           "0x1549487A47332da96cc3EE18D860aaE7b45E2eF3",
				RedirectURIs: []string{"https://example.com/callback"},
				Public:       true,
			},
			expectedError: "",
		},
		{
			name:          "Invalid ClientID Error",
			clientID:      "0xInvalidClientID",
			expectedError: "invalid client ID: 0xInvalidClientID invalid address",
		},
		{
			name:     "Missing Data",
			clientID: "0x1549487A47332da96cc3EE18D860aaE7b45E2eF3",
			mockResponse: map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"message": "No developer license with client id 0x1549487A47332Da96CC3eE18D860aAE7B45e2Ef3.",
						"path": []interface{}{
							"developerLicense",
						},
						"extensions": map[string]interface{}{
							"code": "NOT_FOUND",
						},
					},
				},
				"data": nil,
			},
			mockStatusCode: http.StatusOK,
			expectedClient: storage.Client{},
			expectedError:  "No developer license with client id 0x1549487A47332Da96CC3eE18D860aAE7B45e2Ef3.",
		},
		{
			name:     "Missing Data 1",
			clientID: "0x1549487A47332da96cc3EE18D860aaE7b45E2eF3",
			mockResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"developerLicense": nil,
				},
			},
			mockStatusCode: http.StatusOK,
			expectedClient: storage.Client{},
			expectedError:  "developer license not found for client ID: 0x1549487A47332da96cc3EE18D860aaE7b45E2eF3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.mockResponse != nil {
					w.WriteHeader(tt.mockStatusCode)
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Set environment variable
			os.Setenv("IDENTITY_API_URL", server.URL)
			defer os.Unsetenv("IDENTITY_API_URL")

			// Call the function
			client, err := getDeveloperClient(tt.clientID)

			// Validate results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedClient, client)
			}
		})
	}
}
