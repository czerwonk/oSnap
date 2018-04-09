package config

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/czerwonk/testutils/assert"
)

func TestParseConfig(t *testing.T) {
	b, err := ioutil.ReadFile("tests/config_test.yml")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	assert.StringEqual("cluster", "my-cluster", cfg.Cluster, t)
	assert.IntEqual("keep", 3, cfg.Keep, t)

	assert.StringEqual("api.url", "https://my-ovirt.net", cfg.API.URL, t)
	assert.StringEqual("api.username", "my-osnap-user", cfg.API.User, t)
	assert.StringEqual("api.password", "my-pass", cfg.API.Password, t)
	assert.True("api.insecure", cfg.API.Insecure, t)

	includes := []string{"web.*", "app.*"}
	if !reflect.DeepEqual(includes, cfg.Includes) {
		t.Fatalf("expected includes %v, but got %v", includes, cfg.Includes)
	}

	excludes := []string{"db.*", "temp.*"}
	if !reflect.DeepEqual(excludes, cfg.Excludes) {
		t.Fatalf("expected excludes %v, but got %v", excludes, cfg.Excludes)
	}
}
