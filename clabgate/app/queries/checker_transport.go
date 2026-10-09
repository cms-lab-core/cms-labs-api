package queries

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	corev1 "k8s.io/api/core/v1"
)

// The checker emits one repeated log record prefix followed by the same
// full-result SHA-256 and one base64 chunk. No termination message or
// BEGIN/END marker is needed to reconstruct and verify the entire grade.
const checkerResultPrefix = "CMS_LABS_CHECKER_RESULT_V1 "
const maxCheckerResultBytes = 1024 * 1024
const maxCheckerLogReadBytes = 2 * 1024 * 1024
const maxCheckerDiagnosticBytes = 64 * 1024
const maxCheckerLogTailLines = 2048

// readCheckerPodResult is shared by live CHECK status and CMS grade sync.
func (k *KubernetesAdminQuery) readCheckerPodResult(
	ctx context.Context, namespace, podName string,
) (payload, diagnostics string, err error) {
	limit := int64(maxCheckerLogReadBytes)
	tailLines := int64(maxCheckerLogTailLines)
	raw, readErr := k.clientset.CoreV1().Pods(namespace).GetLogs(
		podName, &corev1.PodLogOptions{
			Container:  "checker",
			LimitBytes: &limit,
			TailLines:  &tailLines,
		},
	).DoRaw(ctx)
	if readErr != nil {
		return "", "", fmt.Errorf("read checker result from Pod logs: %w", readErr)
	}
	return decodeCheckerPodLogFrame(raw)
}

// decodeCheckerPodLogFrame rejects missing, truncated, duplicated, corrupted,
// or oversized grades. Each line repeats the full-payload checksum; no
// positional BEGIN/END marker is necessary. The digest protects integrity,
// not against deliberate forgery from an untrusted container image.
func decodeCheckerPodLogFrame(raw []byte) (payload, diagnostics string, err error) {
	const maxEncodedBytes = (maxCheckerResultBytes + 2) / 3 * 4
	var encoded strings.Builder
	var clean strings.Builder
	var checksum string
	lines := strings.Split(string(raw), "\n")
	// Collect diagnostics before decoding so crashes and malformed frames still
	// expose ordinary logs, including messages following the damaged record.
	for _, line := range lines {
		if !strings.HasPrefix(line, "CMS_LABS_CHECKER_RESULT_") {
			clean.WriteString(line)
			clean.WriteByte('\n')
		}
	}
	diagnostics = tailCheckerDiagnostics(clean.String())
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		if strings.HasPrefix(line, checkerResultPrefix) {
			digest, chunk, ok := strings.Cut(strings.TrimPrefix(line, checkerResultPrefix), " ")
			if !ok || len(digest) != 64 || len(chunk) == 0 || len(chunk) > 2048 {
				return "", diagnostics, errors.New("invalid checker result record")
			}
			if _, decodeErr := hex.DecodeString(digest); decodeErr != nil {
				return "", diagnostics, errors.New("invalid checker result checksum")
			}
			if checksum != "" && digest != checksum {
				return "", diagnostics, errors.New("inconsistent checker result checksums")
			}
			if encoded.Len()+len(chunk) > maxEncodedBytes {
				return "", diagnostics, errors.New("checker result exceeds size limit")
			}
			checksum = digest
			encoded.WriteString(chunk)
			continue
		}
		if strings.HasPrefix(line, "CMS_LABS_CHECKER_RESULT_") {
			return "", diagnostics, errors.New("unrecognized checker result record")
		}
	}
	if encoded.Len() == 0 {
		return "", diagnostics, errors.New("checker result missing from Pod logs")
	}
	result, decodeErr := base64.StdEncoding.DecodeString(encoded.String())
	if decodeErr != nil {
		return "", diagnostics, fmt.Errorf("decode checker result: %w", decodeErr)
	}
	if len(result) == 0 || len(result) > maxCheckerResultBytes {
		return "", diagnostics, errors.New("checker result size is out of bounds")
	}
	sum := sha256.Sum256(result)
	if hex.EncodeToString(sum[:]) != checksum {
		return "", diagnostics, errors.New("checker result incomplete or corrupted (checksum mismatch)")
	}
	if !json.Valid(result) {
		return "", diagnostics, errors.New("checker result is not valid JSON")
	}
	return string(result), diagnostics, nil
}

func tailCheckerDiagnostics(logs string) string {
	if len(logs) > maxCheckerDiagnosticBytes {
		logs = logs[len(logs)-maxCheckerDiagnosticBytes:]
		for len(logs) > 0 && !utf8.RuneStart(logs[0]) {
			logs = logs[1:]
		}
	}
	return strings.TrimSpace(logs)
}
