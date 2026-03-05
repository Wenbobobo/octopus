package common

import "testing"

func TestMessageChanDropOldest(t *testing.T) {
	ch := NewMessageChanWithPolicy(2, QueuePolicyDropOldest)
	in := ch.In()

	for i := 1; i <= 20; i++ {
		in <- &OctopusEvent{ID: Itoa(int64(i))}
	}
	close(in)

	var got []string
	for event := range ch.Out() {
		got = append(got, event.ID)
	}

	if len(got) == 0 {
		t.Fatalf("expected events, got none")
	}
	if ch.Dropped() == 0 {
		t.Fatalf("expected dropped events, got 0")
	}
	if got[len(got)-1] != "20" {
		t.Fatalf("expected last event to be newest item 20, got %s (%v)", got[len(got)-1], got)
	}
}
