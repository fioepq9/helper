package helper_test

import (
	"github.com/google/uuid"
	"github.com/onsi/gomega/gmeasure"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fioepq9/helper"
)

var _ = Describe("viper", Label("viper"), func() {
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
	for i := 0; i < 20000; i++ {
		os.Setenv(uuid.New().String(), uuid.NewString())
	}

	It("unmarshal success", func() {
		err := helper.Viper().Unmarshal(&c)
		Expect(err).To(BeNil())
		Expect(c.Foo).To(Equal("bar"))
		Expect(c.File.Hello).To(Equal("world"))
		Expect(c.File.Question).To(Equal([]string{"what", "is", "the", "answer"}))
		Expect(c.Timeout).To(Equal(10 * time.Second))
		Expect(c.People).To(Equal([]string{"alice", "bob", "carol"}))
		Expect(c.Date.Unix()).To(Equal(int64(1546272000)))
		Expect(c.When.Unix()).To(Equal(int64(1675371906)))
		Expect(c.File2).To(Equal(map[string]any{
			"hello":    "world",
			"question": []any{"what", "is", "the", "answer"},
		}))
		Expect(c.File3).To(Equal([]struct {
			Name string `yaml:"name"`
			Age  int    `yaml:"age"`
		}{
			{"alice", 18},
			{"bob", 19},
			{"carol", 20},
		}))
	})

	It("bench", Serial, func() {
		experiment := gmeasure.NewExperiment("viper - Benchmark")
		AddReportEntry(experiment.Name, experiment)
		experiment.Sample(func(idx int) {
			experiment.MeasureDuration("env", func() {
				_ = helper.Viper(func(viperHelper *helper.ViperHelper) {
					viperHelper.ConfigFile = ""
				}).Unmarshal(&c)
			})
		}, gmeasure.SamplingConfig{N: 200, Duration: time.Minute})
	})
})
