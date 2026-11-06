package save

import (
	"os"
	"testing"
)

func FuzzReadNeverPanicsOnArbitraryJSON(f *testing.F) {
	f.Add([]byte("{"))
	f.Add([]byte(`{"schema_version":1,"faction":"runner"}`))
	f.Add([]byte(`{"schema_version":1,"faction":"corporation","corp":{"capital":-1}}`))
	f.Fuzz(func(t *testing.T, payload []byte) {
		dir := t.TempDir()
		if err := os.WriteFile(Path(dir), payload, 0600); err != nil {
			t.Fatal(err)
		}
		_, _ = Read(dir)
	})
}
