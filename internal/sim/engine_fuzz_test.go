package sim

import (
	"reflect"
	"testing"

	"github.com/jappoman/netrise/internal/game"
)

func FuzzUnknownCommandIsAtomic(f *testing.F) {
	f.Add("not-a-command", int64(0), int64(0))
	f.Add("", int64(-1), int64(99))
	f.Fuzz(func(t *testing.T, kind string, ticks, quantity int64) {
		if kind == string(game.AdvanceTicks) || kind == string(game.CollectContract) {
			t.Skip()
		}
		e := New(catalog(), 401)
		before := NewState(401)
		got, events, err := e.Apply(before, game.Command{Kind: game.CommandKind(kind), Ticks: ticks, Quantity: quantity})
		if err == nil {
			t.Fatalf("unknown command accepted: %q", kind)
		}
		if len(events) != 0 || !reflect.DeepEqual(got, before) {
			t.Fatalf("failed command mutated state: kind=%q events=%#v got=%#v before=%#v", kind, events, got, before)
		}
	})
}
