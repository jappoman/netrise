package content

import (
	"strings"
	"testing"
)

func TestLoadRejectsDuplicateIDs(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"generators":[{"id":"x"},{"id":"x"}]}`))
	if err == nil {
		t.Fatal("expected duplicate rejection")
	}
}

func TestValidateRejectsResearchCycles(t *testing.T) {
	_, err := Load(strings.NewReader(`{"schema_version":1,"research":[{"id":"a","name":"A","prerequisites":["b"]},{"id":"b","name":"B","prerequisites":["a"]}]}`))
	if err == nil {
		t.Fatal("research cycle accepted")
	}
}
