package agentsdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"connectrpc.com/connect"

	agentv1 "github.com/abcp-sdk/agent-proto/agent/v1"
	"github.com/abcp-sdk/agent-proto/agent/v1/agentv1connect"
)

// stubAgent implements just the surface the SDK tests exercise; every other
// method stays Unimplemented.
type stubAgent struct {
	agentv1connect.UnimplementedAgentServiceHandler
	sessions []string
}

func (s *stubAgent) Health(
	ctx context.Context, req *connect.Request[agentv1.HealthRequest],
) (*connect.Response[agentv1.HealthResponse], error) {
	return connect.NewResponse(&agentv1.HealthResponse{Ok: true, Name: "stub-agent"}), nil
}

func (s *stubAgent) ListSessions(
	ctx context.Context, req *connect.Request[agentv1.ListSessionsRequest],
) (*connect.Response[agentv1.ListSessionsResponse], error) {
	res := &agentv1.ListSessionsResponse{}
	for _, name := range s.sessions {
		res.Sessions = append(res.Sessions, &agentv1.Session{Name: name})
	}
	return connect.NewResponse(res), nil
}

func (s *stubAgent) CreateSession(
	ctx context.Context, req *connect.Request[agentv1.CreateSessionRequest],
) (*connect.Response[agentv1.CreateSessionResponse], error) {
	for _, name := range s.sessions {
		if name == req.Msg.GetName() {
			return nil, connect.NewError(connect.CodeAlreadyExists,
				errString("session exists"))
		}
	}
	s.sessions = append(s.sessions, req.Msg.GetName())
	return connect.NewResponse(&agentv1.CreateSessionResponse{}), nil
}

func errString(s string) error { return &simpleError{s} }

type simpleError struct{ msg string }

func (e *simpleError) Error() string { return e.msg }

// recorder captures what actually hit the wire: protocol version + auth.
type recorder struct {
	proto atomic.Value
	auth  atomic.Value
}

// newH2CTestServer serves the stub agent over cleartext HTTP/2 (prior
// knowledge) AND HTTP/1.1 on the same listener, mirroring how the real
// agent deployment behaves for h2c/h1 clients.
func newH2CTestServer(t *testing.T, stub *stubAgent) (*httptest.Server, *recorder) {
	t.Helper()
	rec := &recorder{}
	mux := http.NewServeMux()
	mux.Handle(agentv1connect.NewAgentServiceHandler(stub))
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.proto.Store(r.Proto)
		rec.auth.Store(r.Header.Get("Authorization"))
		mux.ServeHTTP(w, r)
	}))
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	srv.Config.Protocols = protocols
	srv.Start()
	t.Cleanup(srv.Close)
	return srv, rec
}

// TestNewSpeaksH2C proves the default client dials cleartext HTTP/2 (prior
// knowledge) — the agent serves HTTP/2 only, so this is the contract.
func TestNewSpeaksH2C(t *testing.T) {
	srv, rec := newH2CTestServer(t, &stubAgent{sessions: []string{"alpha", "beta"}})

	client := New(srv.URL, "")
	sessions, err := client.ListSessions(context.Background())
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 2 || !sessions["alpha"] || !sessions["beta"] {
		t.Fatalf("sessions = %v", sessions)
	}
	if got := rec.proto.Load(); got != "HTTP/2.0" {
		t.Fatalf("protocol = %v, want HTTP/2.0 (h2c prior knowledge)", got)
	}
	if got := rec.auth.Load(); got != "Bearer devtoken" {
		t.Fatalf("authorization = %v, want Bearer devtoken (default)", got)
	}
}

// TestWithHTTPClientOverride proves the option wiring: a plain HTTP/1.1
// client is honored (and the test server still serves it).
func TestWithHTTPClientOverride(t *testing.T) {
	srv, rec := newH2CTestServer(t, &stubAgent{})

	client := New(srv.URL, "tok-1", WithHTTPClient(http.DefaultClient))
	if _, err := client.ListSessions(context.Background()); err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if got := rec.proto.Load(); got != "HTTP/1.0" && got != "HTTP/1.1" {
		t.Fatalf("protocol = %v, want HTTP/1.x (overridden client)", got)
	}
	if got := rec.auth.Load(); got != "Bearer tok-1" {
		t.Fatalf("authorization = %v, want Bearer tok-1", got)
	}
}

// TestEnsureSessionIdempotent: AlreadyExists maps to nil.
func TestEnsureSessionIdempotent(t *testing.T) {
	stub := &stubAgent{}
	srv, _ := newH2CTestServer(t, stub)
	client := New(srv.URL, "")

	if err := client.EnsureSession(context.Background(), "gamma"); err != nil {
		t.Fatalf("first EnsureSession: %v", err)
	}
	if err := client.EnsureSession(context.Background(), "gamma"); err != nil {
		t.Fatalf("second EnsureSession must be idempotent, got %v", err)
	}
	if len(stub.sessions) != 1 {
		t.Fatalf("stub sessions = %v, want exactly one", stub.sessions)
	}
}
