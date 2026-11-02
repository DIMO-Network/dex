package server

import "github.com/dexidp/dex/server/tokens"

// Issuer is the server's token issuer. The gRPC API needs it to sign DIMO's custom tokens.
func (s *Server) Issuer() *tokens.Issuer { return s.issuer }
