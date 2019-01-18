package config

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
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

	expected := &Config{
		Cluster: "my-cluster",
		Keep:    3,
		API: &APIConfig{
			URL:      "https://my-ovirt.net",
			User:     "my-osnap-user",
			Password: "my-pass",
			Insecure: true,
		},
		Includes: []string{"web.*", "app.*"},
		Excludes: []string{"db.*", "temp.*"},
	}

	assert.Equal(t, expected, cfg)
}
