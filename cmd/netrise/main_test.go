package main

import (
	"path/filepath"
	"testing"
)

func TestCatalogBundlePathsAreExecutableRelative(t *testing.T) {
	paths := catalogBundlePaths(filepath.Join("release", "netrise.exe"))
	want := []string{
		filepath.Join("release", "assets", "content", "catalog.json"),
		filepath.Join("release", "catalog.json"),
	}
	if len(paths) != len(want) {
		t.Fatalf("candidate count = %d, want %d", len(paths), len(want))
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Fatalf("candidate %d = %q, want %q", i, paths[i], want[i])
		}
	}
}
