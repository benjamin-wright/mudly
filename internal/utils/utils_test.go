package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type getwdResponse struct {
	wd  string
	err error
}

type osMock struct {
	getwdResponse getwdResponse
	timestamp     int64
}

func (o *osMock) Getwd() (string, error) { return o.getwdResponse.wd, o.getwdResponse.err }
func (o *osMock) GetTimestamp() int64    { return o.timestamp }

func TestUpdateConfig(t *testing.T) {
	for _, test := range []struct {
		Name     string
		Data     TimestampData
		Config   string
		Artefact string
		Step     string
		Expected *TimestampData
		Error    string
	}{
		{
			Name:     "add",
			Data:     TimestampData{},
			Config:   ".",
			Artefact: "my-artefact",
			Step:     "my-step",
			Expected: &TimestampData{
				Configs: []Config{
					{
						Path: "/base/dir",
						Artefacts: []Artefact{
							{
								Name: "my-artefact",
								Steps: []Step{
									{
										Name:      "my-step",
										Timestamp: 1234,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "update",
			Data: TimestampData{
				Configs: []Config{
					{
						Path: "/base/dir",
						Artefacts: []Artefact{
							{
								Name: "my-artefact",
								Steps: []Step{
									{
										Name:      "my-step",
										Timestamp: 4321,
									},
								},
							},
						},
					},
				},
			},
			Config:   ".",
			Artefact: "my-artefact",
			Step:     "my-step",
			Expected: &TimestampData{
				Configs: []Config{
					{
						Path: "/base/dir",
						Artefacts: []Artefact{
							{
								Name: "my-artefact",
								Steps: []Step{
									{
										Name:      "my-step",
										Timestamp: 1234,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "add new config",
			Data: TimestampData{
				Configs: []Config{
					{
						Path: "/base/dir/other-conf",
						Artefacts: []Artefact{
							{
								Name: "other-artefact",
								Steps: []Step{
									{
										Name:      "other-step",
										Timestamp: 4321,
									},
								},
							},
						},
					},
				},
			},
			Config:   ".",
			Artefact: "my-artefact",
			Step:     "my-step",
			Expected: &TimestampData{
				Configs: []Config{
					{
						Path: "/base/dir/other-conf",
						Artefacts: []Artefact{
							{
								Name: "other-artefact",
								Steps: []Step{
									{
										Name:      "other-step",
										Timestamp: 4321,
									},
								},
							},
						},
					},
					{
						Path: "/base/dir",
						Artefacts: []Artefact{
							{
								Name: "my-artefact",
								Steps: []Step{
									{
										Name:      "my-step",
										Timestamp: 1234,
									},
								},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(test.Name, func(u *testing.T) {
			mock := osMock{getwdResponse: getwdResponse{wd: "/base/dir"}, timestamp: 1234}
			osInstance = &mock

			updated, err := updateConfig(test.Data, test.Config, test.Artefact, test.Step)

			if test.Error != "" {
				assert.EqualError(u, err, test.Error)
			} else {
				assert.NoError(u, err)
			}

			if test.Expected != nil {
				assert.Equal(u, test.Expected, &updated)
			}
		})
	}
}
