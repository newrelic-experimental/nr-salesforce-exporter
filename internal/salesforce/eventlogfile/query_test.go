package eventlogfile

import (
	"newrelic/multienv/pkg/config"
	"regexp"
	"strings"
	"testing"
	"time"
)

var (
	queryRE regexp.Regexp
)

func init() {
	queryRE = *regexp.MustCompile(`^SELECT Id,EventType,CreatedDate,LogDate,Interval,LogFile,Sequence FROM EventLogFile WHERE CreatedDate>=[\d]{4}-[\d]{2}-[\d]{2}T[\d]{2}:[\d]{2}:[\d]{2}\.[\d]{3}Z AND CreatedDate<[\d]{4}-[\d]{2}-[\d]{2}T[\d]{2}:[\d]{2}:[\d]{2}\.[\d]{3}Z AND Interval='Hourly'$`)
}

func TestMakeQuery(t *testing.T) {
	t.Run(
		"returns correct query",
		func(t *testing.T) {
			d, _ := time.ParseDuration("5m")

			url := makeQuery(time.Now(), d, kCreatedDate, kHourly)

			if !queryRE.Match([]byte(url)) {
				t.Errorf("url does not match expected pattern: %s", url)
				return
			}
		},
	)
}

func TestParseDateField(t *testing.T) {
	t.Run(
		"returns error if missing dateField",
		func(t *testing.T) {
			pipeConfig := &config.PipelineConfig{}

			_, err := parseDateField(pipeConfig)
			if err == nil {
				t.Errorf("err was nil: expected missing date field")
				return
			}

			if !strings.Contains(err.Error(), "missing") {
				t.Errorf("unexpected error message: %s", err.Error())
				return
			}
		},
	)

	t.Run(
		"returns kCreatedDate for CreatedDate",
		func(t *testing.T) {
			pipeConfig := &config.PipelineConfig{}
			pipeConfig.Custom = make(map[string]any)
			createdDates := []string{string(kCreatedDate), "cReAteDDATE"}

			for _, d := range createdDates {
				pipeConfig.Custom["dateField"] = d

				createdDate, err := parseDateField(pipeConfig)
				if err != nil {
					t.Errorf("unexpected error parsing %s: %v", d, err)
					return
				}

				if kCreatedDate != createdDate {
					t.Errorf("unexpected value parsing CreatedDate: CreatedDate != %s", createdDate)
					return
				}
			}
		},
	)

	t.Run(
		"returns kLogDate for LogDate",
		func(t *testing.T) {
			pipeConfig := &config.PipelineConfig{}
			pipeConfig.Custom = make(map[string]any)
			logDates := []string{string(kLogDate), "LOgDAte"}

			for _, d := range logDates {
				pipeConfig.Custom["dateField"] = d

				logDate, err := parseDateField(pipeConfig)
				if err != nil {
					t.Errorf("unexpected error parsing %s: %v", d, err)
					return
				}

				if kLogDate != logDate {
					t.Errorf("unexpected value parsing LogDate: LogDate != %s", logDate)
					return
				}
			}
		},
	)

	t.Run(
		"returns error for invalid date",
		func(t *testing.T) {
			pipeConfig := &config.PipelineConfig{}
			pipeConfig.Custom = make(map[string]any)

			pipeConfig.Custom["dateField"] = "Invalid dateField"

			_, err := parseDateField(pipeConfig)
			if err == nil {
				t.Errorf("err was nil: expected invalid date field")
				return
			}

			if !strings.Contains(err.Error(), "invalid") {
				t.Errorf("unexpected error message: %s", err.Error())
				return
			}
		},
	)
}

func TestParseGenerationInterval(t *testing.T) {
	t.Run(
		"returns error if missing generationInterval",
		func(t *testing.T) {
			pipeConfig := &config.PipelineConfig{}

			_, err := parseGenerationInterval(pipeConfig)
			if err == nil {
				t.Errorf("err was nil: expected missing generation interval")
				return
			}

			if !strings.Contains(err.Error(), "missing") {
				t.Errorf("unexpected error message: %s", err.Error())
				return
			}
		},
	)

	t.Run(
		"returns kHourly for Hourly",
		func(t *testing.T) {
			pipeConfig := &config.PipelineConfig{}
			pipeConfig.Custom = make(map[string]any)
			intervals := []string{string(kHourly), "HOUrLY"}

			for _, g := range intervals {
				pipeConfig.Custom["generationInterval"] = g

				generationInterval, err := parseGenerationInterval(pipeConfig)
				if err != nil {
					t.Errorf("unexpected error parsing %s: %v", g, err)
					return
				}

				if kHourly != generationInterval {
					t.Errorf("unexpected value parsing Hourly: Hourly != %s", generationInterval)
					return
				}
			}
		},
	)

	t.Run(
		"returns kDaily for Daily",
		func(t *testing.T) {
			pipeConfig := &config.PipelineConfig{}
			pipeConfig.Custom = make(map[string]any)
			intervals := []string{string(kDaily), "DaILy"}

			for _, g := range intervals {
				pipeConfig.Custom["generationInterval"] = g

				generationInterval, err := parseGenerationInterval(pipeConfig)
				if err != nil {
					t.Errorf("unexpected error parsing %s: %v", g, err)
					return
				}

				if kDaily != generationInterval {
					t.Errorf("unexpected value parsing LogDate: Daily != %s", generationInterval)
					return
				}
			}
		},
	)

	t.Run(
		"returns error for invalid generation interval",
		func(t *testing.T) {
			pipeConfig := &config.PipelineConfig{}
			pipeConfig.Custom = make(map[string]any)

			pipeConfig.Custom["generationInterval"] = "Invalid generationInterval"

			_, err := parseGenerationInterval(pipeConfig)
			if err == nil {
				t.Errorf("err was nil: expected invalid generation interval")
				return
			}

			if !strings.Contains(err.Error(), "invalid") {
				t.Errorf("unexpected error message: %s", err.Error())
				return
			}
		},
	)
}

func TestFetch(t *testing.T) {
	t.Run(
		"fails if api.Query() does",
		func(t *testing.T) {

		},
	)
}
