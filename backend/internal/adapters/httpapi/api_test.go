package httpapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kenar/backend/internal/adapters/memory"
	"github.com/kenar/backend/internal/app/auth"
	"github.com/kenar/backend/internal/app/pairing"
	"github.com/kenar/backend/internal/app/widgets"
	"github.com/kenar/backend/internal/platform/i18n"
	"github.com/kenar/backend/internal/platform/logger"
)

type captureSMS struct{ body string }

func (c *captureSMS) Send(_ context.Context, _, message string) error {
	c.body = message
	return nil
}

type harness struct {
	srv *httptest.Server
	sms *captureSMS
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	bundle, err := i18n.Load("../../../i18n", "fa")
	if err != nil {
		t.Fatalf("load i18n: %v", err)
	}
	users := memory.NewUsers()
	pairs := memory.NewPairs()
	bus := memory.NewBus()
	sms := &captureSMS{}

	router := NewRouter(Deps{
		Logger:  logger.New(),
		I18n:    bundle,
		Auth:    auth.New(users, memory.NewOTPs(), memory.NewSessions(time.Now), sms, bundle, time.Now),
		Pairing: pairing.New(users, memory.NewInvites(), pairs, time.Now),
		Widgets: widgets.New(memory.NewWidgets(), pairs, bus, time.Now),
		Devices: memory.NewDevices(),
		Events:  bus,
	})
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return &harness{srv: srv, sms: sms}
}

// do issues a JSON request, optionally authenticated, and decodes the response.
func (h *harness) do(t *testing.T, method, path, token string, body any) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, h.srv.URL+path, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	out := map[string]any{}
	raw, _ := io.ReadAll(resp.Body)
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return resp.StatusCode, out
}

func (h *harness) login(t *testing.T, phone string) string {
	t.Helper()
	if status, _ := h.do(t, "POST", "/v1/auth/otp/request", "", map[string]string{"phone": phone}); status != 200 {
		t.Fatalf("otp request status = %d", status)
	}
	code := h.sms.body[len(h.sms.body)-6:]
	status, body := h.do(t, "POST", "/v1/auth/otp/verify", "", map[string]string{"phone": phone, "code": code})
	if status != 200 {
		t.Fatalf("otp verify status = %d", status)
	}
	tok, _ := body["token"].(string)
	if tok == "" {
		t.Fatalf("no token in verify response: %v", body)
	}
	return tok
}

func TestFullPairingAndWidgetFlow(t *testing.T) {
	h := newHarness(t)

	alice := h.login(t, "09120000001")
	bob := h.login(t, "09120000002")

	// Unauthenticated access is rejected.
	if status, _ := h.do(t, "GET", "/v1/space", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("unauth space status = %d, want 401", status)
	}

	// Alice creates an invite; Bob accepts it.
	status, inv := h.do(t, "POST", "/v1/invites", alice, nil)
	if status != http.StatusCreated {
		t.Fatalf("create invite status = %d", status)
	}
	code, _ := inv["code"].(string)
	if code == "" {
		t.Fatalf("no invite code: %v", inv)
	}
	if status, _ := h.do(t, "POST", "/v1/invites/accept", bob, map[string]string{"code": code}); status != 200 {
		t.Fatalf("accept invite status = %d", status)
	}

	// Both now see the shared space with each other as partner.
	status, space := h.do(t, "GET", "/v1/space", alice, nil)
	if status != 200 || space["partner_id"] == "" {
		t.Fatalf("space = %d %v", status, space)
	}

	// Alice sets her mood; Bob reads it back.
	payload := base64.StdEncoding.EncodeToString([]byte("mood:happy"))
	status, set := h.do(t, "POST", "/v1/widgets/mood", alice, map[string]any{"payload": payload})
	if status != 200 || set["version"].(float64) != 1 {
		t.Fatalf("set widget = %d %v", status, set)
	}

	status, got := h.do(t, "GET", "/v1/widgets/mood", bob, nil)
	if status != 200 {
		t.Fatalf("get widget status = %d", status)
	}
	states, _ := got["states"].([]any)
	if len(states) != 1 {
		t.Fatalf("states = %v", got)
	}
	first := states[0].(map[string]any)
	if first["payload"] != payload {
		t.Fatalf("payload roundtrip mismatch: %v", first["payload"])
	}

	// Cannot accept your own invite (fresh user, fresh invite).
	carol := h.login(t, "09120000003")
	_, inv2 := h.do(t, "POST", "/v1/invites", carol, nil)
	if status, _ := h.do(t, "POST", "/v1/invites/accept", carol, map[string]string{"code": inv2["code"].(string)}); status != http.StatusBadRequest {
		t.Fatalf("self-accept status = %d, want 400", status)
	}
}

func TestStreamDeliversWidgetEvent(t *testing.T) {
	h := newHarness(t)
	alice := h.login(t, "09120000001")
	bob := h.login(t, "09120000002")

	_, inv := h.do(t, "POST", "/v1/invites", alice, nil)
	h.do(t, "POST", "/v1/invites/accept", bob, map[string]string{"code": inv["code"].(string)})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", h.srv.URL+"/v1/stream", nil)
	req.Header.Set("Authorization", "Bearer "+bob)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("stream status = %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	// Read until the ": connected" comment, confirming the subscription is live.
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("reading connected line: %v", err)
		}
		if strings.Contains(line, "connected") {
			break
		}
	}

	dataLine := make(chan string, 1)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if strings.HasPrefix(line, "data:") {
				dataLine <- line
				return
			}
		}
	}()

	// Alice acts; Bob's stream should receive the widget event.
	payload := base64.StdEncoding.EncodeToString([]byte("mood:loving"))
	if status, _ := h.do(t, "POST", "/v1/widgets/mood", alice, map[string]any{"payload": payload}); status != 200 {
		t.Fatalf("set widget status = %d", status)
	}

	select {
	case line := <-dataLine:
		if !strings.Contains(line, "\"widget_kind\":\"mood\"") {
			t.Fatalf("unexpected event data: %q", line)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("did not receive widget event over stream")
	}
}
