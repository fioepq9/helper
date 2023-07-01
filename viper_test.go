package helper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestViper(t *testing.T) {
	type Config struct {
		Foo  string `yaml:"foo"`
		File struct {
			Hello    string   `yaml:"hello"`
			Question []string `yaml:"question"`
		} `yaml:"file"`
		Timeout time.Duration  `yaml:"timeout"`
		People  []string       `yaml:"people"`
		Date    time.Time      `yaml:"date"`
		When    time.Time      `yaml:"when"`
		File2   map[string]any `yaml:"file2"`
		File3   []struct {
			Name string `yaml:"name"`
			Age  int    `yaml:"age"`
		}
	}

	var c Config
	err := Viper().Unmarshal(&c)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "bar", c.Foo)
	assert.Equal(t, "world", c.File.Hello)
	assert.Equal(t, []string{"what", "is", "the", "answer"}, c.File.Question)
	assert.Equal(t, 10*time.Second, c.Timeout)
	assert.Equal(t, []string{"alice", "bob", "carol"}, c.People)
	assert.Equal(t, int64(1546272000), c.Date.Unix())
	assert.Equal(t, int64(1675371906), c.When.Unix())
	assert.Equal(t, map[string]any{
		"hello":    "world",
		"question": []any{"what", "is", "the", "answer"},
	}, c.File2)
	assert.Equal(t, []struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}{
		{"alice", 18},
		{"bob", 19},
		{"carol", 20},
	}, c.File3)
}
