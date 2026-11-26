package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fiatjaf/eventstore/lmdb"
	"github.com/nbd-wtf/go-nostr"
)

// TestLmdbCountMatchesFilter checks lmdb's CountEvents and CountEventsHLL
// against nostr.Filter.Matches, for the filter shapes a relay is handed: no
// constraint, one index, an index plus extra author, kind or tag checks, and a
// time window. Each call has to come back within a deadline.
func TestLmdbCountMatchesFilter(t *testing.T) {
	db := newTestDB(t, 500).(*lmdb.LMDBBackend)
	// CountEventsHLL calls this without a nil check; haven itself never sets it
	db.EnableHLLCacheFor = func(int) (bool, bool) { return false, false }

	alice, bob := nostr.GeneratePrivateKey(), nostr.GeneratePrivateKey()
	alicePub, _ := nostr.GetPublicKey(alice)
	bobPub, _ := nostr.GetPublicKey(bob)
	target := "a1" + strings.Repeat("0", 62) // an event id that is only ever referenced

	var events []*nostr.Event
	add := func(sk string, kind int, ts nostr.Timestamp, tags nostr.Tags) {
		evt := &nostr.Event{CreatedAt: ts, Kind: kind, Tags: tags, Content: "x"}
		if err := evt.Sign(sk); err != nil {
			t.Fatalf("sign: %v", err)
		}
		if err := db.SaveEvent(context.Background(), evt); err != nil {
			t.Fatalf("save: %v", err)
		}
		events = append(events, evt)
	}
	for i := 0; i < 5; i++ {
		ts := nostr.Timestamp(1700000000 + i)
		add(alice, nostr.KindTextNote, ts, nostr.Tags{{"t", "haven"}})
		add(bob, nostr.KindTextNote, ts, nostr.Tags{{"p", alicePub}})
		add(bob, nostr.KindReaction, ts, nostr.Tags{{"e", target}, {"p", alicePub}})
	}
	add(alice, nostr.KindFollowList, 1700000100, nostr.Tags{{"p", bobPub}})

	filters := map[string]nostr.Filter{
		"empty":                {},
		"kind":                 {Kinds: []int{nostr.KindTextNote}},
		"two kinds":            {Kinds: []int{nostr.KindTextNote, nostr.KindReaction}},
		"author":               {Authors: []string{bobPub}},
		"author and kind":      {Authors: []string{bobPub}, Kinds: []int{nostr.KindReaction}},
		"two authors and kind": {Authors: []string{alicePub, bobPub}, Kinds: []int{nostr.KindTextNote}},
		"tag":                  {Tags: nostr.TagMap{"p": {alicePub}}},
		"kind and tag":         {Kinds: []int{nostr.KindTextNote}, Tags: nostr.TagMap{"p": {alicePub}}},
		"author kind and tag":  {Authors: []string{alicePub, bobPub}, Kinds: []int{nostr.KindTextNote}, Tags: nostr.TagMap{"t": {"haven"}}},
		"since":                {Kinds: []int{nostr.KindTextNote}, Since: ptr(nostr.Timestamp(1700000003))},
		"id":                   {IDs: []string{events[0].ID}},
		"no match":             {Kinds: []int{30023}},
	}

	for name, filter := range filters {
		var want int64
		for _, evt := range events {
			if filter.Matches(evt) {
				want++
			}
		}

		got := countWithDeadline(t, name, func() (int64, error) { return db.CountEvents(context.Background(), filter) })
		if got != want {
			t.Errorf("%s: CountEvents = %d, want %d", name, got, want)
		}
	}

	// the shapes NIP-45 answers with a HyperLogLog: reactions to an event, followers
	hll := map[string]nostr.Filter{
		"reactions": {Kinds: []int{nostr.KindReaction}, Tags: nostr.TagMap{"e": {target}}},
		"followers": {Kinds: []int{nostr.KindFollowList}, Tags: nostr.TagMap{"p": {bobPub}}},
	}
	for name, filter := range hll {
		var want int64
		for _, evt := range events {
			if filter.Matches(evt) {
				want++
			}
		}

		got := countWithDeadline(t, name+" (hll)", func() (int64, error) {
			n, _, err := db.CountEventsHLL(context.Background(), filter, 0)
			return n, err
		})
		if got != want {
			t.Errorf("%s: CountEventsHLL = %d, want %d", name, got, want)
		}
	}
}

func countWithDeadline(t *testing.T, name string, count func() (int64, error)) int64 {
	t.Helper()
	type result struct {
		n   int64
		err error
	}
	done := make(chan result, 1)
	go func() {
		n, err := count()
		done <- result{n, err}
	}()
	select {
	case r := <-done:
		if r.err != nil {
			t.Fatalf("%s: %v", name, r.err)
		}
		return r.n
	case <-time.After(10 * time.Second):
		t.Fatalf("%s: the count did not return within 10s", name)
		return 0
	}
}

func ptr[T any](v T) *T { return &v }
