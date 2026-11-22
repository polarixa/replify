package replify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fullConfig returns a config with every field set to a non-default value.
func fullConfig() *SignatureConfig {
	return NewSignatureConfig("secret-a").
		WithAlgorithm(HMACSHA512).
		WithMaxAge(7*time.Minute).
		DisableIncludeTimestamp().
		EnableCanonicalize().
		WithHeaders("X-One", "X-Two").
		WithIgnoringSignatureFields("sig1", "sig2").
		WithIgnoringResponseFields("resp1", "resp2")
}

// signBody computes the base64 HMAC-SHA256 of the payload exactly as VerifySignature expects it.
func signBody(secret string, body []byte, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(buildSignaturePayload(body, ts, nil, ""))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// T1: Clone must copy every field and must not share slices with the original.
func TestSignatureConfig_CloneCopiesEveryFieldAndIsDeep(t *testing.T) {
	t.Parallel()

	orig := fullConfig()
	clone := orig.Clone()

	if clone.SecretKey() != "secret-a" || clone.Algorithm() != HMACSHA512 ||
		clone.MaxAge() != 7*time.Minute || clone.IsIncludeTimestamp() || !clone.IsCanonicalize() {
		t.Fatalf("scalar fields not cloned: %+v", clone)
	}
	if !slices.Equal(clone.HeadersToSign(), []string{"X-One", "X-Two"}) {
		t.Errorf("headersToSign = %v", clone.HeadersToSign())
	}
	if !slices.Equal(clone.IgnoringSignatureFields(), []string{"sig1", "sig2"}) {
		t.Errorf("ignoringSignatureFields dropped or wrong: %v", clone.IgnoringSignatureFields())
	}
	if !slices.Equal(clone.IgnoringResponseFields(), []string{"resp1", "resp2"}) {
		t.Errorf("ignoringResponseFields dropped or wrong: %v", clone.IgnoringResponseFields())
	}

	if len(clone.headersToSign) != 2 || len(clone.ignoringSignatureFields) != 2 || len(clone.ignoringResponseFields) != 2 {
		t.Fatalf("clone is missing list entries: headers=%d sig=%d resp=%d",
			len(clone.headersToSign), len(clone.ignoringSignatureFields), len(clone.ignoringResponseFields))
	}

	// Mutate the clone's slices in place and by append: the original must not change.
	clone.headersToSign[0] = "MUTATED"
	clone.ignoringSignatureFields[0] = "MUTATED"
	clone.ignoringResponseFields[0] = "MUTATED"
	clone.WithIgnoringSignatureFields("extra")
	clone.WithIgnoringResponseFields("extra")

	if !slices.Equal(orig.headersToSign, []string{"X-One", "X-Two"}) {
		t.Errorf("original headersToSign changed through the clone: %v", orig.headersToSign)
	}
	if !slices.Equal(orig.ignoringSignatureFields, []string{"sig1", "sig2"}) {
		t.Errorf("original ignoringSignatureFields changed through the clone: %v", orig.ignoringSignatureFields)
	}
	if !slices.Equal(orig.ignoringResponseFields, []string{"resp1", "resp2"}) {
		t.Errorf("original ignoringResponseFields changed through the clone: %v", orig.ignoringResponseFields)
	}
}

func TestSignatureConfig_CloneNil(t *testing.T) {
	t.Parallel()
	var c *SignatureConfig
	if c.Clone() != nil {
		t.Error("Clone of nil must be nil")
	}
}

// T2: signature.Clone must not share the headers map.
func TestSignature_CloneIsDeep(t *testing.T) {
	t.Parallel()

	orig := NewSignature().WithValue("v").WithTimeUnix(42).WithHeader("A", "1").WithHeader("B", "2")
	clone := orig.Clone()

	if clone.Value() != "v" || clone.Timestamp() != 42 || clone.LenHeaders() != 2 {
		t.Fatalf("clone fields wrong: %+v", clone)
	}
	clone.headers["A"] = "MUTATED"
	clone.headers["C"] = "3"

	if orig.headers["A"] != "1" || len(orig.headers) != 2 {
		t.Errorf("original headers changed through the clone: %v", orig.headers)
	}
}

// T3: Equal must compare every field unconditionally.
func TestSignatureConfig_EqualTruthTable(t *testing.T) {
	t.Parallel()

	if !fullConfig().Equal(fullConfig()) {
		t.Error("identical configs must be equal")
	}
	var nilA, nilB *SignatureConfig
	if !nilA.Equal(nilB) {
		t.Error("nil/nil must be equal")
	}
	if nilA.Equal(fullConfig()) || fullConfig().Equal(nilA) {
		t.Error("nil/non-nil must not be equal")
	}

	// Each pair differs in exactly one field; none may compare equal.
	rows := []struct {
		name   string
		mutate func(c *SignatureConfig)
	}{
		{"secret set vs empty", func(c *SignatureConfig) { c.secretKey = "" }},
		{"secret differs", func(c *SignatureConfig) { c.secretKey = "other" }},
		{"algorithm set vs empty", func(c *SignatureConfig) { c.algorithm = "" }},
		{"algorithm differs", func(c *SignatureConfig) { c.algorithm = HMACSHA256 }},
		{"maxAge set vs zero", func(c *SignatureConfig) { c.maxAge = 0 }},
		{"maxAge differs", func(c *SignatureConfig) { c.maxAge = time.Hour }},
		{"headers list vs empty", func(c *SignatureConfig) { c.headersToSign = nil }},
		{"headers differ", func(c *SignatureConfig) { c.headersToSign = []string{"X-One", "X-Other"} }},
		{"canonicalize true vs false", func(c *SignatureConfig) { c.canonicalize = false }},
		{"includeTimestamp differs", func(c *SignatureConfig) { c.includeTimestamp = true }},
		{"ignoringSignatureFields differ", func(c *SignatureConfig) { c.ignoringSignatureFields = []string{"other"} }},
		{"ignoringSignatureFields vs empty", func(c *SignatureConfig) { c.ignoringSignatureFields = nil }},
		{"ignoringResponseFields differ", func(c *SignatureConfig) { c.ignoringResponseFields = []string{"other"} }},
		{"ignoringResponseFields vs empty", func(c *SignatureConfig) { c.ignoringResponseFields = nil }},
	}
	for _, row := range rows {
		a, b := fullConfig(), fullConfig()
		row.mutate(b)
		if a.Equal(b) || b.Equal(a) {
			t.Errorf("%s: configs differing only in this field compared equal", row.name)
		}
	}

	// nil and empty slices are equal.
	a, b := fullConfig(), fullConfig()
	a.ignoringResponseFields, b.ignoringResponseFields = nil, []string{}
	if !a.Equal(b) {
		t.Error("nil and empty slices must compare equal")
	}
}

// T4: IsExpired is deprecated and never reports expiry; it must still be nil-safe.
func TestSignatureConfig_IsExpiredIsDeterministicFalse(t *testing.T) {
	t.Parallel()

	var nilCfg *SignatureConfig
	if nilCfg.IsExpired() {
		t.Error("nil config must not report expired")
	}
	if NewSignatureConfig("k").WithMaxAge(time.Nanosecond).IsExpired() {
		t.Error("IsExpired carries no timestamp and must always be false")
	}
	if NewSignatureConfig("k").IsExpired() {
		t.Error("IsExpired must be false for a default config")
	}
}

// T5: future-dated signatures beyond the clock skew must be rejected.
func TestSignatureVerify_FutureTimestamps(t *testing.T) {
	t.Parallel()

	const secret = "future-secret"
	body := []byte(`{"a":1}`)
	cfg := NewSignatureConfig(secret) // maxAge 5m, default skew 30s

	verifyAt := func(ts int64) bool {
		provided := NewSignature().WithAlgorithm(HMACSHA256).WithTimeUnix(ts).WithValue(signBody(secret, body, ts))
		ok, _ := VerifySignature(cfg, body, provided)
		return ok
	}
	now := time.Now()

	tests := []struct {
		name string
		ts   int64
		want bool
	}{
		{"now", now.Unix(), true},
		{"now+10s within skew", now.Add(10 * time.Second).Unix(), true},
		{"now+2m beyond skew", now.Add(2 * time.Minute).Unix(), false},
		{"now+1h", now.Add(time.Hour).Unix(), false},
		{"now+10y", now.Add(10 * 365 * 24 * time.Hour).Unix(), false},
		{"past beyond maxAge", now.Add(-10 * time.Minute).Unix(), false},
	}
	for _, tt := range tests {
		if got := verifyAt(tt.ts); got != tt.want {
			t.Errorf("%s: VerifySignature = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSignatureConfig_ClockSkewAPI(t *testing.T) {
	t.Parallel()

	if DefaultSignatureMaxClockSkew != 30*time.Second {
		t.Errorf("DefaultSignatureMaxClockSkew = %v, want 30s", DefaultSignatureMaxClockSkew)
	}
	if got := NewSignatureConfig("k").MaxClockSkew(); got != DefaultSignatureMaxClockSkew {
		t.Errorf("default skew = %v, want %v", got, DefaultSignatureMaxClockSkew)
	}
	if got := NewSignatureConfig("k").WithMaxClockSkew(5 * time.Second).MaxClockSkew(); got != 5*time.Second {
		t.Errorf("WithMaxClockSkew not applied: %v", got)
	}
	if got := (&SignatureConfig{}).MaxClockSkew(); got != 0 {
		t.Errorf("zero-value config skew = %v, want 0 (strict)", got)
	}
	// Strict skew: any future time is expired.
	strict := NewSignatureConfig("k").WithMaxClockSkew(0)
	if !strict.IsExpiredTime(time.Now().Add(time.Minute)) {
		t.Error("with skew 0 a future time must be reported expired")
	}
	// Within skew keeps the full max age.
	if got := NewSignatureConfig("k").RemainingMaxAge(time.Now().Add(5 * time.Second)); got != 5*time.Minute {
		t.Errorf("within skew RemainingMaxAge = %v, want maxAge", got)
	}
}

// T6: legacy algorithms are rejected unless explicitly enabled.
func TestSignatureLegacyAlgorithmsGate(t *testing.T) {
	t.Parallel()

	body := []byte(`{"a":1}`)
	for _, algo := range []SignatureAlgorithm{HMACSHA1, HMACMD5} {
		t.Run(algo.String(), func(t *testing.T) {
			t.Parallel()

			cfg := NewSignatureConfig("legacy-secret").WithAlgorithm(algo)
			if cfg.IsLegacyAlgorithmsEnabled() {
				t.Fatal("legacy algorithms must be disabled by default")
			}

			if s, w := GenerateSignature(cfg, body); s != nil || w.StatusCode() != http.StatusBadRequest {
				t.Errorf("GenerateSignature by default: signature=%v status=%d, want nil and 400", s, w.StatusCode())
			}
			probe := NewSignature().WithAlgorithm(algo).WithTimeUnix(time.Now().Unix()).WithValue("x")
			if ok, w := VerifySignature(cfg, body, probe); ok || w.StatusCode() != http.StatusBadRequest {
				t.Errorf("VerifySignature by default: ok=%v status=%d, want false and 400", ok, w.StatusCode())
			}
			if _, w := GenerateSignatureFromRequest(httptest.NewRequest(http.MethodGet, "/", nil), cfg, body); w.StatusCode() != http.StatusBadRequest {
				t.Errorf("GenerateSignatureFromRequest by default: status=%d, want 400", w.StatusCode())
			}

			cfg.EnableLegacyAlgorithms()
			if !cfg.IsLegacyAlgorithmsEnabled() {
				t.Fatal("EnableLegacyAlgorithms did not enable legacy algorithms")
			}
			s, w := GenerateSignature(cfg, body)
			if s == nil || w.IsError() {
				t.Fatalf("GenerateSignature after opt-in failed: %v", w.Error())
			}
			if ok, w := VerifySignature(cfg, body, s); !ok {
				t.Fatalf("round trip after opt-in failed: %v", w.Message())
			}
		})
	}

	if NewSignatureConfig("k").WithLegacyAlgorithms(true).WithLegacyAlgorithms(false).IsLegacyAlgorithmsEnabled() {
		t.Error("WithLegacyAlgorithms(false) must disable legacy algorithms")
	}
}

// Clone, Equal and Respond must carry the new fields.
func TestSignatureConfig_NewFieldsFlowThroughCloneEqualRespond(t *testing.T) {
	t.Parallel()

	c := NewSignatureConfig("k").WithMaxClockSkew(9 * time.Second).EnableLegacyAlgorithms()
	clone := c.Clone()
	if clone.MaxClockSkew() != 9*time.Second || !clone.IsLegacyAlgorithmsEnabled() {
		t.Errorf("clone lost new fields: skew=%v legacy=%v", clone.MaxClockSkew(), clone.IsLegacyAlgorithmsEnabled())
	}
	if !c.Equal(clone) {
		t.Error("config must equal its clone")
	}
	if c.Equal(NewSignatureConfig("k").EnableLegacyAlgorithms()) {
		t.Error("configs with different skew must not be equal")
	}
	if c.Equal(NewSignatureConfig("k").WithMaxClockSkew(9 * time.Second)) {
		t.Error("configs with different legacy setting must not be equal")
	}

	m := c.Respond()
	if m["max_clock_skew"] != (9*time.Second).String() || m["allow_legacy_algorithms"] != true {
		t.Errorf("Respond missing new non-default fields: %v", m)
	}
	def := NewSignatureConfig("k").Respond()
	if _, ok := def["max_clock_skew"]; ok {
		t.Error("default skew must not be included in Respond")
	}
	if _, ok := def["allow_legacy_algorithms"]; ok {
		t.Error("disabled legacy flag must not be included in Respond")
	}
	if m["secret_key"] != "******" {
		t.Errorf("secret must stay masked, got %v", m["secret_key"])
	}
}

// T7: round-trip guards that must hold before and after the fixes.
func TestSignatureRoundTripGuards(t *testing.T) {
	t.Parallel()

	body := []byte(`{"b":2,"a":1}`)

	t.Run("default config", func(t *testing.T) {
		t.Parallel()
		cfg := NewSignatureConfig("rt-secret")
		s, w := GenerateSignature(cfg, body)
		if w.IsError() {
			t.Fatalf("generate: %v", w.Message())
		}
		if ok, w := VerifySignature(cfg, body, s); !ok {
			t.Fatalf("verify: %v", w.Message())
		}
	})
	t.Run("canonicalize on and off", func(t *testing.T) {
		t.Parallel()
		for _, canon := range []bool{true, false} {
			cfg := NewSignatureConfig("rt-secret").WithCanonicalize(canon)
			s, _ := GenerateSignature(cfg, body)
			if ok, w := VerifySignature(cfg, body, s); !ok {
				t.Errorf("canonicalize=%v: %v", canon, w.Message())
			}
		}
		// Canonicalization makes key order irrelevant.
		cfg := NewSignatureConfig("rt-secret").EnableCanonicalize()
		s, _ := GenerateSignature(cfg, body)
		if ok, _ := VerifySignature(cfg, []byte(`{"a":1,"b":2}`), s); !ok {
			t.Error("canonical form must verify regardless of key order")
		}
	})
	t.Run("timestamp disabled", func(t *testing.T) {
		t.Parallel()
		cfg := NewSignatureConfig("rt-secret").DisableIncludeTimestamp()
		s, _ := GenerateSignature(cfg, body)
		if ok, w := VerifySignature(cfg, body, s); !ok {
			t.Fatalf("verify: %v", w.Message())
		}
	})
	t.Run("tampered body", func(t *testing.T) {
		t.Parallel()
		cfg := NewSignatureConfig("rt-secret")
		s, _ := GenerateSignature(cfg, body)
		if ok, _ := VerifySignature(cfg, []byte(`{"b":2,"a":9}`), s); ok {
			t.Error("tampered body must not verify")
		}
	})
	t.Run("algorithm mismatch", func(t *testing.T) {
		t.Parallel()
		s, _ := GenerateSignature(NewSignatureConfig("rt-secret"), body)
		other := NewSignatureConfig("rt-secret").WithAlgorithm(HMACSHA512)
		if ok, _ := VerifySignature(other, body, s); ok {
			t.Error("algorithm mismatch must not verify")
		}
	})
	t.Run("from request with signed headers", func(t *testing.T) {
		t.Parallel()
		cfg := NewSignatureConfig("rt-secret").WithHeaders("X-Tenant")
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X-Tenant", "acme")

		s, w := GenerateSignatureFromRequest(req, cfg, body)
		if w.IsError() {
			t.Fatalf("generate: %v", w.Message())
		}
		if ok, w := VerifySignatureFromRequest(req, cfg, body, s); !ok {
			t.Fatalf("verify: %v", w.Message())
		}
		tampered := httptest.NewRequest(http.MethodPost, "/", nil)
		tampered.Header.Set("X-Tenant", "evil")
		if ok, _ := VerifySignatureFromRequest(tampered, cfg, body, s); ok {
			t.Error("tampered header must not verify")
		}
	})
}

// --- Replay protection ---------------------------------------------------------

// T0 characterization: without replay protection a captured, validly signed request
// verifies every time within the validity window. This is the gap the feature closes.
func TestSignatureReplay_CharacterizationReplayIsPossibleByDefault(t *testing.T) {
	t.Parallel()

	cfg := NewSignatureConfig("replay-secret")
	body := []byte(`{"order":1}`)

	s, w := GenerateSignature(cfg, body)
	if w.IsError() {
		t.Fatalf("generate: %v", w.Message())
	}
	for i := 1; i <= 2; i++ {
		if ok, w := VerifySignature(cfg, body, s); !ok {
			t.Fatalf("verification #%d failed: %v", i, w.Message())
		}
	}
}

// replayConfig returns a config with replay protection enabled and the given store.
func replayConfig(store NonceStore) *SignatureConfig {
	cfg := NewSignatureConfig("replay-secret").EnableReplayProtection()
	if store != nil {
		cfg.WithNonceStore(store)
	}
	return cfg
}

// fakeNonceStore counts calls and can be told to fail or to remember nonces.
type fakeNonceStore struct {
	mu    sync.Mutex
	calls int
	seen  map[string]bool
	err   error
	ttls  []time.Duration
}

func (f *fakeNonceStore) SeenOrMark(nonce string, ttl time.Duration) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.ttls = append(f.ttls, ttl)
	if f.err != nil {
		return false, f.err
	}
	if f.seen == nil {
		f.seen = make(map[string]bool)
	}
	was := f.seen[nonce]
	f.seen[nonce] = true
	return was, nil
}

func (f *fakeNonceStore) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

// T1: round trip with protection enabled; a second verification of the same signature is rejected.
func TestSignatureReplay_ReplayRejected(t *testing.T) {
	t.Parallel()

	cfg := replayConfig(nil)
	body := []byte(`{"order":1}`)

	s, w := GenerateSignature(cfg, body)
	if w.IsError() {
		t.Fatalf("generate: %v", w.Message())
	}
	if len(s.Nonce()) != 32 {
		t.Fatalf("expected a 32-character hex nonce, got %q", s.Nonce())
	}
	if ok, w := VerifySignature(cfg, body, s); !ok {
		t.Fatalf("first verification failed: %v", w.Message())
	}
	ok, w := VerifySignature(cfg, body, s)
	if ok {
		t.Fatal("the replayed signature must be rejected")
	}
	if w.StatusCode() != http.StatusUnauthorized || !strings.Contains(w.Message(), "replay") {
		t.Errorf("expected a 401 naming replay, got %d %q", w.StatusCode(), w.Message())
	}

	// A fresh signature for the same body is a new request and is accepted.
	s2, _ := GenerateSignature(cfg, body)
	if s2.Nonce() == s.Nonce() {
		t.Error("each signature must carry a fresh nonce")
	}
	if ok, w := VerifySignature(cfg, body, s2); !ok {
		t.Errorf("a freshly signed request must verify: %v", w.Message())
	}
}

// T2: with protection enabled, a signature without a nonce is rejected.
func TestSignatureReplay_MissingNonceRejected(t *testing.T) {
	t.Parallel()

	store := &fakeNonceStore{}
	cfg := replayConfig(store)
	body := []byte(`{"order":1}`)

	s, _ := GenerateSignature(cfg, body)
	stripped := s.Clone().WithNonce("")
	if ok, w := VerifySignature(cfg, body, stripped); ok || w.StatusCode() != http.StatusUnauthorized {
		t.Errorf("a signature without a nonce must be rejected with 401, got ok=%v status=%d", ok, w.StatusCode())
	}
	// A signature produced without protection (no nonce at all) is rejected as well.
	plain, _ := GenerateSignature(NewSignatureConfig("replay-secret"), body)
	if ok, _ := VerifySignature(cfg, body, plain); ok {
		t.Error("a nonce-less signature must be rejected when protection is enabled")
	}
	if store.callCount() != 0 {
		t.Errorf("the store must not be touched for these requests, got %d calls", store.callCount())
	}
}

// T3: the HMAC covers the nonce, so swapping it invalidates the signature.
func TestSignatureReplay_TamperedNonceRejected(t *testing.T) {
	t.Parallel()

	store := &fakeNonceStore{}
	cfg := replayConfig(store)
	body := []byte(`{"order":1}`)

	s, _ := GenerateSignature(cfg, body)
	tampered := s.Clone().WithNonce("00000000000000000000000000000000")
	if ok, _ := VerifySignature(cfg, body, tampered); ok {
		t.Fatal("a signature with a swapped nonce must be rejected")
	}
	if store.callCount() != 0 {
		t.Errorf("the store must not be touched when the HMAC fails, got %d calls", store.callCount())
	}
	if ok, w := VerifySignature(cfg, body, s); !ok {
		t.Errorf("the untampered signature must still verify: %v", w.Message())
	}
}

// T4: the memory store expires entries by TTL; expired entries are evicted.
func TestSignatureReplay_MemoryStoreTTL(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	store := newMemoryNonceStore(DefaultNonceStoreCapacity)
	store.now = func() time.Time { return now }

	ttl := 5*time.Minute + 30*time.Second
	if seen, err := store.SeenOrMark("n1", ttl); seen || err != nil {
		t.Fatalf("first use: seen=%v err=%v, want false nil", seen, err)
	}
	if seen, _ := store.SeenOrMark("n1", ttl); !seen {
		t.Fatal("second use inside the window must be seen")
	}

	now = now.Add(ttl - time.Second)
	if seen, _ := store.SeenOrMark("n1", ttl); !seen {
		t.Fatal("just before expiry the nonce must still be seen")
	}

	now = now.Add(2 * time.Second) // past expiry
	if seen, err := store.SeenOrMark("n1", ttl); seen || err != nil {
		t.Fatalf("after expiry the nonce is accepted again: seen=%v err=%v", seen, err)
	}

	// Entries that expired are evicted by later inserts: no leak.
	if _, err := store.SeenOrMark("old", time.Minute); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * ttl)
	if _, err := store.SeenOrMark("fresh", ttl); err != nil {
		t.Fatal(err)
	}
	if got := store.len(); got != 1 {
		t.Errorf("expected only the fresh entry to remain after eviction, got %d entries", got)
	}
}

// TTL passed to the store is maxAge plus the clock skew.
func TestSignatureReplay_StoreTTLIsMaxAgePlusSkew(t *testing.T) {
	t.Parallel()

	store := &fakeNonceStore{}
	cfg := replayConfig(store).WithMaxAge(2 * time.Minute).WithMaxClockSkew(10 * time.Second)
	body := []byte(`{"a":1}`)

	s, _ := GenerateSignature(cfg, body)
	if ok, w := VerifySignature(cfg, body, s); !ok {
		t.Fatalf("verify: %v", w.Message())
	}
	if len(store.ttls) != 1 || store.ttls[0] != 2*time.Minute+10*time.Second {
		t.Errorf("store ttl = %v, want [2m10s]", store.ttls)
	}
}

// T5: nothing reaches the store unless the HMAC (and the earlier checks) pass.
func TestSignatureReplay_StoreNotTouchedBeforeHMAC(t *testing.T) {
	t.Parallel()

	body := []byte(`{"order":1}`)

	t.Run("invalid HMAC", func(t *testing.T) {
		t.Parallel()
		store := &fakeNonceStore{}
		cfg := replayConfig(store)
		s, _ := GenerateSignature(cfg, body)
		if ok, _ := VerifySignature(cfg, []byte(`{"order":2}`), s); ok {
			t.Fatal("a tampered body must be rejected")
		}
		if store.callCount() != 0 {
			t.Errorf("expected zero SeenOrMark calls, got %d", store.callCount())
		}
	})
	t.Run("expired", func(t *testing.T) {
		t.Parallel()
		store := &fakeNonceStore{}
		cfg := replayConfig(store)
		ts := time.Now().Add(-time.Hour).Unix()
		s := NewSignature().WithAlgorithm(HMACSHA256).WithTimeUnix(ts).WithNonce("abc")
		mac := hmac.New(sha256.New, []byte("replay-secret"))
		mac.Write(buildSignaturePayload(body, ts, nil, "abc"))
		s.WithValue(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
		if ok, _ := VerifySignature(cfg, body, s); ok {
			t.Fatal("an expired signature must be rejected")
		}
		if store.callCount() != 0 {
			t.Errorf("expected zero SeenOrMark calls, got %d", store.callCount())
		}
	})
	t.Run("algorithm mismatch", func(t *testing.T) {
		t.Parallel()
		store := &fakeNonceStore{}
		cfg := replayConfig(store)
		s, _ := GenerateSignature(cfg, body)
		s.WithAlgorithm(HMACSHA512)
		if ok, _ := VerifySignature(cfg, body, s); ok {
			t.Fatal("an algorithm mismatch must be rejected")
		}
		if store.callCount() != 0 {
			t.Errorf("expected zero SeenOrMark calls, got %d", store.callCount())
		}
	})
	t.Run("valid signature touches the store exactly once", func(t *testing.T) {
		t.Parallel()
		store := &fakeNonceStore{}
		cfg := replayConfig(store)
		s, _ := GenerateSignature(cfg, body)
		if ok, w := VerifySignature(cfg, body, s); !ok {
			t.Fatalf("verify: %v", w.Message())
		}
		if store.callCount() != 1 {
			t.Errorf("expected one SeenOrMark call, got %d", store.callCount())
		}
	})
}

// T6: a store error fails closed.
func TestSignatureReplay_StoreErrorFailsClosed(t *testing.T) {
	t.Parallel()

	store := &fakeNonceStore{err: errors.New("redis unavailable")}
	cfg := replayConfig(store)
	body := []byte(`{"order":1}`)

	s, _ := GenerateSignature(cfg, body)
	ok, w := VerifySignature(cfg, body, s)
	if ok || w.StatusCode() != http.StatusUnauthorized {
		t.Errorf("a store error must reject with 401, got ok=%v status=%d", ok, w.StatusCode())
	}
}

// T7: with protection disabled the payload bytes are exactly the legacy format.
func TestSignatureReplay_PayloadBackwardCompatible(t *testing.T) {
	t.Parallel()

	body := []byte(`{"a":1}`)
	headers := map[string]string{"X-B": "2", "X-A": "1"}

	tests := []struct {
		name string
		got  []byte
		want string
	}{
		{"body only", buildSignaturePayload(body, 0, nil, ""), `{"a":1}`},
		{"with timestamp", buildSignaturePayload(body, 1700000000, nil, ""), `{"a":1}||ts:1700000000`},
		{"with headers", buildSignaturePayload(body, 1700000000, headers, ""), `{"a":1}||ts:1700000000||h:X-A:1||h:X-B:2`},
		{"with nonce", buildSignaturePayload(body, 1700000000, nil, "abc"), `{"a":1}||ts:1700000000||n:abc`},
	}
	for _, tt := range tests {
		if string(tt.got) != tt.want {
			t.Errorf("%s: payload = %q, want %q", tt.name, tt.got, tt.want)
		}
	}

	// A disabled config signs and verifies the legacy payload, and carries no nonce.
	cfg := NewSignatureConfig("replay-secret")
	s, _ := GenerateSignature(cfg, body)
	if s.Nonce() != "" || s.Respond()["nonce"] != nil {
		t.Errorf("a disabled config must not produce a nonce, got %q", s.Nonce())
	}
	if cfg.IsReplayProtectionEnabled() {
		t.Error("replay protection must be off by default")
	}
	mac := hmac.New(sha256.New, []byte("replay-secret"))
	mac.Write(buildSignaturePayload(body, s.Timestamp(), nil, ""))
	if want := base64.StdEncoding.EncodeToString(mac.Sum(nil)); s.Value() != want {
		t.Error("the disabled signature must equal the HMAC of the legacy payload")
	}
}

// T8: the memory store is bounded and fails closed at capacity.
func TestSignatureReplay_MemoryStoreCapacity(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	store := newMemoryNonceStore(DefaultNonceStoreCapacity)
	store.now = func() time.Time { return now }

	for i := 0; i < DefaultNonceStoreCapacity; i++ {
		if _, err := store.SeenOrMark(strconv.Itoa(i), time.Hour); err != nil {
			t.Fatalf("insert %d failed before capacity: %v", i, err)
		}
	}
	if seen, err := store.SeenOrMark("overflow", time.Hour); err == nil || seen {
		t.Fatalf("a full store of unexpired entries must fail closed, got seen=%v err=%v", seen, err)
	}
	// An already-known nonce is still reported as seen, not as an error.
	if seen, err := store.SeenOrMark("0", time.Hour); !seen || err != nil {
		t.Errorf("a known nonce must be reported as seen, got seen=%v err=%v", seen, err)
	}

	now = now.Add(2 * time.Hour) // every entry expired
	if seen, err := store.SeenOrMark("after-sweep", time.Hour); seen || err != nil {
		t.Fatalf("expired entries must be swept to free space, got seen=%v err=%v", seen, err)
	}
	if store.len() != 1 {
		t.Errorf("expected the sweep to leave only the new entry, got %d", store.len())
	}
}

// T9: concurrent verification of the same fresh signature succeeds exactly once.
func TestSignatureReplay_ConcurrentVerifyExactlyOnce(t *testing.T) {
	t.Parallel()

	cfg := replayConfig(nil)
	body := []byte(`{"order":1}`)
	s, _ := GenerateSignature(cfg, body)

	const workers = 64
	var successes atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if ok, _ := VerifySignature(cfg, body, s); ok {
				successes.Add(1)
			}
		}()
	}
	close(start)
	wg.Wait()

	if got := successes.Load(); got != 1 {
		t.Errorf("expected exactly one successful verification, got %d", got)
	}
}

// T10: the nonce travels in X-Signature-Nonce and the FromRequest path rejects replays.
func TestSignatureReplay_FromRequestAndHeaders(t *testing.T) {
	t.Parallel()

	cfg := replayConfig(nil)
	body := []byte(`{"order":1}`)
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	s, w := GenerateSignatureFromRequest(req, cfg, body)
	if w.IsError() || s.Nonce() == "" {
		t.Fatalf("generate: %v nonce=%q", w.Message(), s.Nonce())
	}

	hdr := http.Header{}
	hdr.Set(HeaderXSignature.String(), s.Value())
	hdr.Set(HeaderXSignatureAlgorithm.String(), s.Algorithm().String())
	hdr.Set(HeaderXSignatureTimestamp.String(), strconv.FormatInt(s.Timestamp(), 10))
	hdr.Set(HeaderXSignatureNonce.String(), s.Nonce())
	if HeaderXSignatureNonce.String() != "X-Signature-Nonce" {
		t.Fatalf("unexpected header name %q", HeaderXSignatureNonce.String())
	}

	provided := SignatureFromHeader(hdr)
	if provided.Nonce() != s.Nonce() {
		t.Fatalf("SignatureFromHeader nonce = %q, want %q", provided.Nonce(), s.Nonce())
	}
	if ok, w := VerifySignatureFromRequest(req, cfg, body, provided); !ok {
		t.Fatalf("first verification failed: %v", w.Message())
	}
	if ok, _ := VerifySignatureFromRequest(req, cfg, body, provided); ok {
		t.Error("the replayed request must be rejected")
	}
}

// ApplySignatureHeader writes the nonce header and the wrapper round trip is single-use.
func TestSignatureReplay_ApplySignatureHeaderRoundTrip(t *testing.T) {
	t.Parallel()

	cfg := replayConfig(nil)
	event := New().OK().WithBody(map[string]any{"invoice": "inv_1"})

	rec := httptest.NewRecorder()
	event.ApplySignatureHeader(rec, cfg)
	if rec.Header().Get(HeaderXSignatureNonce.String()) == "" {
		t.Fatal("ApplySignatureHeader must write the nonce header when protection is enabled")
	}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	for k, v := range rec.Header() {
		req.Header[k] = v
	}
	unwrap := func() *wrapper {
		w, err := UnwrapJSON(event.JSON())
		if err != nil {
			t.Fatalf("unwrap: %v", err)
		}
		return w
	}
	if !unwrap().VerifySignatureHeader(req, cfg) {
		t.Fatal("the first delivery must verify")
	}
	if unwrap().VerifySignatureHeader(req, cfg) {
		t.Error("the replayed delivery must be rejected")
	}
}

// Validation, Clone, Equal, Respond and signature plumbing for the new fields.
func TestSignatureReplay_ConfigAndSignaturePlumbing(t *testing.T) {
	t.Parallel()

	// Enabling requires a bounded window.
	unbounded := NewSignatureConfig("k").EnableReplayProtection().WithMaxAge(0)
	if w := ValidateSignatureConfig(unbounded); !w.IsError() || w.StatusCode() != http.StatusBadRequest {
		t.Errorf("replay protection with maxAge 0 must be a 400, got %d", w.StatusCode())
	}
	if _, w := GenerateSignature(unbounded, []byte(`{}`)); w.StatusCode() != http.StatusBadRequest {
		t.Errorf("GenerateSignature must reject an unbounded window, got %d", w.StatusCode())
	}
	noTimestamp := NewSignatureConfig("k").EnableReplayProtection().DisableIncludeTimestamp()
	if w := ValidateSignatureConfig(noTimestamp); w.StatusCode() != http.StatusBadRequest {
		t.Errorf("replay protection without timestamps must be a 400, got %d", w.StatusCode())
	}
	noStore := NewSignatureConfig("k").EnableReplayProtection().WithNonceStore(nil)
	if w := ValidateSignatureConfig(noStore); w.StatusCode() != http.StatusBadRequest {
		t.Errorf("replay protection without a store must be a 400, got %d", w.StatusCode())
	}
	if w := ValidateSignatureConfig(NewSignatureConfig("k").EnableReplayProtection()); w.IsError() {
		t.Errorf("a valid replay config must validate: %v", w.Message())
	}

	// WithReplayProtection toggles; the default store is created on enable.
	c := NewSignatureConfig("k").WithReplayProtection(true)
	if !c.IsReplayProtectionEnabled() || c.nonceStore == nil {
		t.Error("enabling must create the default store")
	}
	if c.WithReplayProtection(false).IsReplayProtectionEnabled() {
		t.Error("WithReplayProtection(false) must disable")
	}

	// Clone shares the store by reference and copies the flag; Equal compares the flag only.
	store := &fakeNonceStore{}
	orig := replayConfig(store)
	clone := orig.Clone()
	if !clone.IsReplayProtectionEnabled() || clone.nonceStore != NonceStore(store) {
		t.Error("clone must keep the flag and share the store")
	}
	if !orig.Equal(clone) {
		t.Error("a config must equal its clone")
	}
	other := replayConfig(&fakeNonceStore{})
	if !orig.Equal(other) {
		t.Error("the store must be excluded from equality")
	}
	if orig.Equal(NewSignatureConfig("replay-secret")) {
		t.Error("configs differing in replay protection must not be equal")
	}

	// Respond.
	if orig.Respond()["replay_protection"] != true || orig.Respond()["secret_key"] != "******" {
		t.Errorf("Respond = %v", orig.Respond())
	}
	if _, ok := NewSignatureConfig("k").Respond()["replay_protection"]; ok {
		t.Error("default config must not report replay_protection")
	}

	// signature nonce: accessor, Respond, Clone, Equal (presence-XOR).
	a := NewSignature().WithValue("v").WithNonce("n1")
	if a.Nonce() != "n1" || a.Respond()["nonce"] != "n1" {
		t.Errorf("nonce accessor/Respond: %q %v", a.Nonce(), a.Respond())
	}
	if !a.Equal(a.Clone()) || a.Clone().Nonce() != "n1" {
		t.Error("clone must preserve the nonce")
	}
	if a.Equal(NewSignature().WithValue("v")) {
		t.Error("a signature with a nonce must not equal one without")
	}
	if a.Equal(NewSignature().WithValue("v").WithNonce("n2")) {
		t.Error("signatures with different nonces must not be equal")
	}
}
