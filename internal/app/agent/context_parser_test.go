package agent

import (
	"testing"
)

func TestParseContextMentions_SingleFile(t *testing.T) {
	input := "@file path/to/file.txt check this"
	mentions := ParseContextMentions(input)

	if len(mentions) != 1 {
		t.Fatalf("expected 1 mention, got %d", len(mentions))
	}

	if mentions[0].Provider != "file" {
		t.Errorf("expected provider 'file', got '%s'", mentions[0].Provider)
	}
	if mentions[0].Query != "path/to/file.txt" {
		t.Errorf("expected query 'path/to/file.txt', got '%s'", mentions[0].Query)
	}
}

func TestParseContextMentions_MultipleProviders(t *testing.T) {
	input := "@file main.go @git status check the code"
	mentions := ParseContextMentions(input)

	if len(mentions) != 2 {
		t.Fatalf("expected 2 mentions, got %d", len(mentions))
	}

	if mentions[0].Provider != "file" {
		t.Errorf("expected provider 'file', got '%s'", mentions[0].Provider)
	}
	if mentions[1].Provider != "git" {
		t.Errorf("expected provider 'git', got '%s'", mentions[1].Provider)
	}
}

func TestParseContextMentions_NoMentions(t *testing.T) {
	input := "just a normal request without mentions"
	mentions := ParseContextMentions(input)

	if mentions != nil {
		t.Errorf("expected nil, got %v", mentions)
	}
}

func TestStripContextMentions(t *testing.T) {
	input := "@file main.go check the code @git status"
	stripped := StripContextMentions(input)

	expected := "check the code"
	if stripped != expected {
		t.Errorf("expected '%s', got '%s'", expected, stripped)
	}
}

func TestStripContextMentions_NoMentions(t *testing.T) {
	input := "normal request"
	stripped := StripContextMentions(input)

	if stripped != input {
		t.Errorf("expected '%s', got '%s'", input, stripped)
	}
}
