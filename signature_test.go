package replify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"slices"
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
	mac.Write(buildSignaturePayload(body, ts, nil))
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
