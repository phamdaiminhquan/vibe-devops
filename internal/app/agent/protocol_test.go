package agent

import "testing"

func TestExtractFirstJSONObject_Basic(t *testing.T) {
	obj, err := extractFirstJSONObject(`{"type":"done","command":"echo hi"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(obj) != `{"type":"done","command":"echo hi"}` {
		t.Fatalf("unexpected obj: %s", string(obj))
	}
}

func TestExtractFirstJSONObject_Fenced(t *testing.T) {
	obj, err := extractFirstJSONObject("```json\n{\"type\":\"done\",\"command\":\"echo hi\"}\n```")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(obj) != `{"type":"done","command":"echo hi"}` {
		t.Fatalf("unexpected obj: %s", string(obj))
	}
}

func TestParseAction_Done(t *testing.T) {
	a, err := ParseAction(`{"type":"done","command":"echo hi","explanation":"ok"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Type != ActionTypeDone {
		t.Fatalf("expected done")
	}
	if a.Command != "echo hi" {
		t.Fatalf("unexpected command: %q", a.Command)
	}
}

func TestParseAction_Tool_DefaultInput(t *testing.T) {
	a, err := ParseAction(`{"type":"tool","tool":"list_dir"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Type != ActionTypeTool {
		t.Fatalf("expected tool")
	}
	if a.Tool != "list_dir" {
		t.Fatalf("unexpected tool: %q", a.Tool)
	}
	if string(a.Input) != "{}" {
		t.Fatalf("expected default input {} but got %s", string(a.Input))
	}
}

func TestParseAction_Answer(t *testing.T) {
	a, err := ParseAction(`{"type":"answer","explanation":"it failed because ..."}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Type != ActionTypeAnswer {
		t.Fatalf("expected answer")
	}
	if a.Explanation == "" {
		t.Fatalf("expected explanation")
	}
}
