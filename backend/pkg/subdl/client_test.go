package subdl

import (
	"bytes"
	"testing"
)

func TestExtractDownloadedSubtitlePayloadRejectsHTML(t *testing.T) {
	body := []byte("<!DOCTYPE html><html><head><title>Subscene</title></head><body>blocked</body></html>")

	_, _, err := extractDownloadedSubtitlePayload(body)
	if err == nil {
		t.Fatal("expected HTML payload to be rejected")
	}
}

func TestExtractDownloadedSubtitlePayloadAcceptsRawSRT(t *testing.T) {
	body := []byte("1\n00:00:05,000 --> 00:00:06,000\nHey, new wallet?\n")

	data, filename, err := extractDownloadedSubtitlePayload(body)
	if err != nil {
		t.Fatalf("expected raw subtitle text to pass: %v", err)
	}
	if filename != "subtitle.srt" {
		t.Fatalf("filename = %q, want subtitle.srt", filename)
	}
	if !bytes.Equal(data, body) {
		t.Fatal("returned data differs from input body")
	}
}
