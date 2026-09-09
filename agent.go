// Package agentsdk provides the abc AgentService Connect client.
//
// Per the connectrpc convention this package ships ONLY the generated RPC
// client (agentv1connect.AgentServiceClient + agent.v1 message types) plus
// transport/auth primitives:
//   - NewHTTPClient()     — a cleartext HTTP/2 (prior-knowledge) *http.Client
//   - AuthInterceptor()   — bearer auth for connect.NewClient/NewAgentServiceClient
//
// The caller builds its own *http.Client (or uses NewHTTPClient()) and hands it
// to agentv1connect.NewAgentServiceClient(httpClient, baseURL, opts...). Nothing
// here wraps the client in a convenience struct.
package agentsdk

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	agentv1 "github.com/abcp-sdk/agent-proto/agent/v1"
	"github.com/abcp-sdk/agent-proto/agent/v1/agentv1connect"
)

type (
	// Session is an alias for the generated agent.v1.Session.
	Session            = agentv1.Session
	Message            = agentv1.Message
	ListSessionsRequest  = agentv1.ListSessionsRequest
	CreateSessionRequest = agentv1.CreateSessionRequest
	GetSessionRequest    = agentv1.GetSessionRequest
	DeleteSessionRequest = agentv1.DeleteSessionRequest
	ForkRequest          = agentv1.ForkRequest
	GetFileRequest       = agentv1.GetFileRequest
	GetFileMetaRequest   = agentv1.GetFileMetaRequest
)

var (
	// NewAgentServiceClient is the generated Connect client constructor.
	NewAgentServiceClient = agentv1connect.NewAgentServiceClient
)

// NewHTTPClient returns a cleartext-HTTP/2 (prior knowledge) *http.Client,
// matching the agent's HTTP/2-only listener. Pass to NewAgentServiceClient, or
// override for TLS / proxies / timeouts.
func NewHTTPClient() *http.Client {
	protocols := new(http.Protocols)
	protocols.SetHTTP1(false)
	protocols.SetUnencryptedHTTP2(true)
	return &http.Client{
		Transport: &http.Transport{Protocols: protocols},
	}
}

// AuthInterceptor returns a Connect interceptor that attaches
// `Authorization: Bearer <token>` to every request. No header is set when the
// token is empty (caller passed none).
func AuthInterceptor(token string) connect.Interceptor {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if token != "" {
				req.Header().Set("Authorization", "Bearer "+token)
			}
			return next(ctx, req)
		}
	})
}
