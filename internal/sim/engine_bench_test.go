package sim

import "testing"

func BenchmarkAdvanceTenThousandTicks(b *testing.B) {
	catalog := catalog()
	engine := New(catalog, 17)
	state := NewState(17)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.Advance(state, 10000)
	}
}
