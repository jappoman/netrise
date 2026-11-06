package balance

import (
	"reflect"
	"testing"

	"github.com/jappoman/netrise/internal/content"
)

func TestRunProducesDeterministicStrategyMatrix(t *testing.T) {
	catalog := content.Catalog{
		SchemaVersion: 1,
		Targets: []content.Target{{
			ID: "target", Name: "Relay", Reachable: true,
			Layers: []content.Layer{{Integrity: 12, ICE: 2, Counter: 1}}, TimeLimit: 30,
		}},
	}
	first, err := Run(catalog, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Run(catalog, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("balance report was not deterministic:\nfirst=%#v\nsecond=%#v", first, second)
	}
	if first.Attempts != 36 || len(first.Results) != 3 {
		t.Fatalf("unexpected matrix size: %#v", first)
	}
	for _, result := range first.Results {
		if result.Attempts != 12 || result.Victories+result.Defeats+result.Timeouts != result.Attempts || result.AverageDurationTicks <= 0 {
			t.Fatalf("invalid strategy summary: %#v", result)
		}
	}
}

func TestRunRejectsOversizedSeedRange(t *testing.T) {
	if _, err := Run(content.Catalog{}, 1, MaxSeedCount+1); err == nil {
		t.Fatal("oversized balance range was accepted")
	}
}

func TestRunHandlesMaximumSeedWithoutLoopOverflow(t *testing.T) {
	catalog := content.Catalog{
		Targets: []content.Target{{ID: "target", Reachable: true, Layers: []content.Layer{{Integrity: 1}}, TimeLimit: 5}},
	}
	const maximumInt64 = int64(9223372036854775807)
	report, err := Run(catalog, maximumInt64, maximumInt64)
	if err != nil || report.Attempts != 12 {
		t.Fatalf("maximum seed did not resolve safely: report=%#v err=%v", report, err)
	}
}
