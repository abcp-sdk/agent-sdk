// Package agentsdk re-exports the buf-generated abc AgentService Connect client
// and agent.v1 message types.
//
// Nothing here is hand-written: for the RPC client itself import the generated
// package directly:
//
//	agentv1 "github.com/abcp-sdk/agent-proto/agent/v1"
//	"github.com/abcp-sdk/agent-proto/agent/v1/agentv1connect"
//
//	client := agentv1connect.NewAgentServiceClient(httpClient, baseURL, opts...)
//
// This package exists only so consumers have the common message/request names
// and the client constructor under one import.
package agentsdk

import (
	agentv1 "github.com/abcp-sdk/agent-proto/agent/v1"
	"github.com/abcp-sdk/agent-proto/agent/v1/agentv1connect"
)

type (
	// Session is an alias for the generated agent.v1.Session.
	Session              = agentv1.Session
	Message              = agentv1.Message
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
