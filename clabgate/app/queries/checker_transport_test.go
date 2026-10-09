package queries

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func makeResultLog(payload string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(payload))
	hash := sha256.Sum256([]byte(payload))
	digest := hex.EncodeToString(hash[:])
	var result strings.Builder
	for pos := 0; pos < len(encoded); pos += 2048 {
		end := pos + 2048
		if end > len(encoded) {
			end = len(encoded)
		}
		result.WriteString(checkerResultPrefix + digest + " " + encoded[pos:end] + "\n")
	}
	return result.String()
}

func TestDecodeCheckerPodLogFrameOver4KiBAndDiagnostics(t *testing.T) {
	message := strings.Repeat("MAC learned; ", 8000)
	payload := fmt.Sprintf(`{"max_score":12,"current_score":12,"result_display":"12/12","tasks":[{"title":"FDB","logs":[{"message":%q}],"complete":true}]}`, message)
	raw := []byte("normal log before\n" + makeResultLog(payload) + "normal log after\n")
	result, diagnostics, err := decodeCheckerPodLogFrame(raw)
	if err != nil {
		t.Fatalf("decodeCheckerPodLogFrame(): %v", err)
	}
	if result != payload || len(result) <= 4096 {
		t.Fatalf("payload changed or never exceeded 4KiB: %d bytes", len(result))
	}
	if !strings.Contains(diagnostics, "normal log before") || !strings.Contains(diagnostics, "normal log after") {
		t.Fatalf("diagnostic logs were not preserved: %q", diagnostics)
	}
	if strings.Contains(diagnostics, checkerResultPrefix) {
		t.Fatal("result records leaked into student-visible logs")
	}
}

func TestDecodeCheckerPodLogFrameIgnoresLargeDiagnosticPrefix(t *testing.T) {
	payload := `{"max_score":1,"current_score":1,"result_display":"1/1"}`
	logs := []byte(strings.Repeat("log entry\n", 12000) + makeResultLog(payload))
	got, diagnostic, err := decodeCheckerPodLogFrame(logs)
	if err != nil || got != payload {
		t.Fatalf("large logs: payload=%q, err=%v", got, err)
	}
	if len(diagnostic) > maxCheckerDiagnosticBytes {
		t.Fatalf("diagnostic logs exceed cap: %d", len(diagnostic))
	}
}

func TestDecodeCheckerPodLogFrameRejectsInvalidMessages(t *testing.T) {
	payload := `{"max_score":1,"current_score":1,"result_display":"1/1"}`
	valid := makeResultLog(payload)
	prefix := checkerResultPrefix
	tamperedHash := strings.Repeat("0", 64)
	if strings.Contains(valid, tamperedHash) {
		t.Fatal("unexpected test checksum")
	}
	tests := []struct {
		name string
		body string
	}{
		{"missing frame", "ordinary logs only\n"},
		{"missing first record", strings.Join(strings.Split(makeResultLog(strings.Repeat("x", 7000)), "\n")[1:], "\n")},
		{"duplicate record", valid + valid},
		{"checksum mismatch", strings.Replace(valid, valid[len(prefix):len(prefix)+64], tamperedHash, 1)},
		{"invalid hex", strings.Replace(valid, valid[len(prefix):len(prefix)+64], strings.Repeat("z", 64), 1)},
		{"bad base64", strings.Replace(valid, prefix, prefix+"! ", 1)},
		{"partial log", valid[:len(valid)/2]},
		{"oversized chunk", prefix + tamperedHash + " " + strings.Repeat("x", 2049) + "\n"},
		{"oversized frame", strings.Repeat(prefix+tamperedHash+" "+strings.Repeat("x", 2048)+"\n", 700)},
		{"unknown record", "CMS_LABS_CHECKER_RESULT_WRONG " + valid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _, err := decodeCheckerPodLogFrame([]byte(tt.body)); err == nil {
				t.Fatalf("accepted invalid frame, %d payload bytes", len(got))
			}
		})
	}
}

func TestDecodeCheckerPodLogFrameRejectsMissingMiddleRecord(t *testing.T) {
	payload := fmt.Sprintf(`{"report":%q}`, strings.Repeat("payload; ", 800))
	lines := strings.Split(strings.TrimSpace(makeResultLog(payload)), "\n")
	if len(lines) < 3 {
		t.Fatal("test needs multiple chunks")
	}
	truncated := strings.Join(append(lines[:1], lines[2:]...), "\n")
	if _, _, err := decodeCheckerPodLogFrame([]byte(truncated)); err == nil {
		t.Fatal("accepted result with missing interior record")
	}
}

func TestInvalidResultPreservesDiagnostics(t *testing.T) {
	for _, frame := range []string{"", "CMS_LABS_CHECKER_RESULT_V1 broken\n", "CMS_LABS_CHECKER_RESULT_UNKNOWN data\n"} {
		_, logs, err := decodeCheckerPodLogFrame([]byte("before\n" + frame + "checker failed: SSH timeout\n"))
		if err == nil || !strings.Contains(logs, "before") || !strings.Contains(logs, "SSH timeout") {
			t.Fatalf("missing crash diagnostics: logs=%q error=%v", logs, err)
		}
		if strings.Contains(logs, "CMS_LABS_CHECKER_RESULT_") {
			t.Fatal("protocol records leaked into diagnostics")
		}
	}
}

func TestDiagnosticTailPreservesUTF8(t *testing.T) {
	logs := tailCheckerDiagnostics(strings.Repeat("я", maxCheckerDiagnosticBytes) + "x")
	if !utf8.ValidString(logs) || len(logs) > maxCheckerDiagnosticBytes {
		t.Fatal("diagnostic tail is oversized or cuts a UTF-8 character")
	}
}

func TestMaximumResultFitsLogBudget(t *testing.T) {
	base := `{"max_score":1,"current_score":1,"result_display":"1/1","report":""}`
	payload := strings.Replace(base, `"report":""`, `"report":"`+strings.Repeat("x", maxCheckerResultBytes-len(base))+`"`, 1)
	frame := makeResultLog(payload)
	if len(frame) > maxCheckerLogReadBytes || strings.Count(frame, "\n") > maxCheckerLogTailLines {
		t.Fatal("maximum result cannot be retrieved within Pod log budgets")
	}
	got, _, err := decodeCheckerPodLogFrame([]byte(frame))
	if err != nil || got != payload {
		t.Fatalf("maximum result rejected: %v", err)
	}
}
