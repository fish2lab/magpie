package sessions

import (
	"path/filepath"
	"strings"
	"testing"
)

// A fork subagent's file opens with a fork-context-ref line, then a copy of
// the parent's reply that started the fork, then the fork's own prompt: the
// copy is the parent's call, counted once, in the parent's file. The shapes
// are Claude Code 2.1's, their ids and words replaced.
func TestClaudeForkSubagentCountsParentReplyOnce(t *testing.T) {
	inZone(t, 0)
	claude, _ := setup(t)
	id := "33333333-2222-3333-4444-555555555555"
	dir := filepath.Join(claude, "projects", "-work-night")
	sidechain := func(line string) string {
		return strings.Replace(line, `"isSidechain":false`, `"isSidechain":true,"agentId":"f1"`, 1)
	}
	writeFile(t, filepath.Join(dir, id+".jsonl"), ccMsg("p1", "2026-09-21T10:00:00.000Z", 100, 10))
	writeFile(t, filepath.Join(dir, id, "subagents", "agent-f1.jsonl"),
		`{"type":"fork-context-ref","agentId":"f1","parentSessionId":"`+id+`","parentLastUuid":"u1","contextLength":1}`+"\n"+
			sidechain(ccMsg("p1", "2026-09-21T10:00:00.000Z", 100, 10))+
			`{"parentUuid":"u2","isSidechain":true,"agentId":"f1","type":"user","message":{"role":"user","content":"port it"},"timestamp":"2026-09-21T10:00:01.000Z","cwd":"/work/night","sessionId":"`+id+`"}`+"\n"+
			sidechain(ccMsg("f2", "2026-09-21T10:00:05.000Z", 50, 5)))

	want := Tokens{Input: 150, Output: 15, CacheRead: 20}
	if s := find(t, List(0), "claude", id); s.Tokens != want {
		t.Fatalf("session: %+v, want %+v", s.Tokens, want)
	}
	if got, _ := usageOn(statsAt(0, statsNow), "2026-09-21", "/work/night"); got != want {
		t.Fatalf("stats: %+v, want %+v", got, want)
	}
}
