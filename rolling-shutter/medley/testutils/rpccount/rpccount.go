// Package rpccount provides a JSON-RPC counting proxy in front of an
// ethclient/simulated backend. Tests use it to observe the number of RPC
// calls a piece of code issues, without touching production code.
//
// The proxy stands up an httptest.Server, forwards every incoming JSON-RPC
// request to the simulated backend's in-process *rpc.Client, and records
// which methods were invoked. Tests interact with the backend through a
// regular *ethclient.Client obtained via ethclient.Dial against the proxy
// URL, so no code under test needs to be aware of the counting layer.
package rpccount

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/rpc"
)

// Backend bundles a simulated blockchain, a counting HTTP proxy in front of
// it, and an *ethclient.Client that talks to the proxy.
type Backend struct {
	// Sim is the underlying simulated blockchain. Use Sim.Commit() to seal
	// blocks, Sim.AdjustTime, etc.
	Sim *simulated.Backend
	// Client is a real *ethclient.Client whose transport goes through the
	// counting proxy.
	Client *ethclient.Client
	// URL is the http endpoint of the counting proxy.
	URL string

	proxy    *httptest.Server
	upstream *rpc.Client

	mu     sync.Mutex
	counts map[string]int
}

// NewBackend creates a simulated backend, wraps it with a counting HTTP
// JSON-RPC proxy, and returns a Backend ready to hand out an *ethclient.Client.
// Cleanup is registered on t.
func NewBackend(t *testing.T, alloc types.GenesisAlloc) *Backend {
	t.Helper()

	sim := simulated.NewBackend(alloc, simulated.WithBlockGasLimit(8_000_000))

	simC := sim.Client()
	// simulated.simClient embeds an *ethclient.Client in an anonymous field
	// named "Client" (after the type name). That field name shadows
	// ethclient.Client's Client() *rpc.Client method, so an interface
	// assertion won't work. Reach it by reflection instead.
	v := reflect.ValueOf(simC)
	ecField := v.FieldByName("Client")
	if !ecField.IsValid() {
		t.Fatalf("rpccount: simulated.Client concrete type %T has no embedded *ethclient.Client field", simC)
	}
	ec, ok := ecField.Interface().(*ethclient.Client)
	if !ok {
		t.Fatalf("rpccount: embedded Client field is %T, not *ethclient.Client", ecField.Interface())
	}
	upstream := ec.Client()

	b := &Backend{
		Sim:      sim,
		upstream: upstream,
		counts:   map[string]int{},
	}
	b.proxy = httptest.NewServer(http.HandlerFunc(b.serve))
	b.URL = b.proxy.URL

	client, err := ethclient.Dial(b.URL)
	if err != nil {
		b.proxy.Close()
		sim.Close()
		t.Fatalf("rpccount: ethclient.Dial(%s): %v", b.URL, err)
	}
	b.Client = client

	t.Cleanup(func() {
		client.Close()
		b.proxy.Close()
		_ = sim.Close()
	})
	return b
}

// Counts returns a snapshot of method -> call count. Method names are the
// JSON-RPC method fields, e.g. "eth_getLogs", "eth_blockNumber",
// "eth_getBlockByNumber".
func (b *Backend) Counts() map[string]int {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make(map[string]int, len(b.counts))
	for k, v := range b.counts {
		out[k] = v
	}
	return out
}

// Reset clears the counts. Useful between phases of a test.
func (b *Backend) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.counts = map[string]int{}
}

// LogCounts writes the current counts to t.Log in a stable, comma-separated
// form: "method=count, method=count, ..." with methods sorted.
func (b *Backend) LogCounts(t *testing.T, label string) {
	t.Helper()
	counts := b.Counts()
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf []byte
	for i, k := range keys {
		if i > 0 {
			buf = append(buf, ',', ' ')
		}
		buf = append(buf, k...)
		buf = append(buf, '=')
		buf = append(buf, []byte(itoa(counts[k]))...)
	}
	t.Logf("%s: %s", label, string(buf))
}

// itoa is a tiny int-to-string helper to avoid importing strconv purely for a
// log line.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}

type rpcRequest struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      json.RawMessage   `json:"id,omitempty"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (b *Backend) serve(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Detect batch vs single by peeking at first non-whitespace byte.
	first := firstJSONByte(body)
	if first == '[' {
		var batch []rpcRequest
		if err := json.Unmarshal(body, &batch); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		responses := make([]rpcResponse, len(batch))
		for i, req := range batch {
			responses[i] = b.dispatch(r.Context(), req)
		}
		writeJSON(w, responses)
		return
	}

	var req rpcRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, b.dispatch(r.Context(), req))
}

func (b *Backend) dispatch(ctx context.Context, req rpcRequest) rpcResponse {
	b.mu.Lock()
	b.counts[req.Method]++
	b.mu.Unlock()

	// Forward params as []interface{} of json.RawMessage. rpc.Client marshals
	// []interface{} into a params array; each json.RawMessage marshals as its
	// verbatim bytes, so the upstream call sees exactly the original params.
	params := make([]interface{}, len(req.Params))
	for i, p := range req.Params {
		params[i] = p
	}
	var raw json.RawMessage
	if err := b.upstream.CallContext(ctx, &raw, req.Method, params...); err != nil {
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32000, Message: err.Error()},
		}
	}
	return rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: raw}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func firstJSONByte(body []byte) byte {
	for _, c := range body {
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			continue
		}
		return c
	}
	return 0
}
