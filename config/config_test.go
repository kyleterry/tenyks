package config

import (
	"os"
	"testing"
)

func TestDiscoverConfig(t *testing.T) {
	test_path1 := "/tmp/catsandboots/tenyks_config.json"
	test_path2 := "/tmp/tmp_tenyks_config.json"

	f, err := os.Create(test_path2)
	if err != nil {
		t.Error("Could not create file for testing: ", test_path2)
	}
	f.Close()

	ConfigSearch.AddPath(test_path1)
	ConfigSearch.AddPath(test_path2)
	path := discoverConfig()
	if path != test_path2 {
		t.Error("Expected", test_path2, "got", path)
	}
}
