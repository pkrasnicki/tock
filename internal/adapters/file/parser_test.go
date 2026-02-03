package file_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kriuchkov/tock/internal/adapters/file"
	"github.com/kriuchkov/tock/internal/core/models"
)

func TestParseActivity(t *testing.T) {
	localTime := func(year int, month time.Month, day, hour, minute, sec int) time.Time {
		return time.Date(year, month, day, hour, minute, sec, 0, time.Local)
	}

	tests := []struct {
		name      string
		line      string
		want      *models.Activity
		wantErr   bool
		errTarget error
	}{
		{
			name: "valid activity with start and end time",
			line: "2023-10-27 10:00 - 2023-10-27 11:00 | Project A | Working on feature X",
			want: &models.Activity{
				StartTime:   localTime(2023, 10, 27, 10, 0, 0),
				EndTime:     func() *time.Time { t := localTime(2023, 10, 27, 11, 0, 0); return &t }(),
				Project:     "Project A",
				Description: "Working on feature X",
				Attributes:  []models.Attribute{},
			},
			wantErr: false,
		},
		{
			name: "valid activity with start time only",
			line: "2023-10-27 12:00 | Project B | Meeting",
			want: &models.Activity{
				StartTime:   localTime(2023, 10, 27, 12, 0, 0),
				EndTime:     nil,
				Project:     "Project B",
				Description: "Meeting",
				Attributes:  []models.Attribute{},
			},
			wantErr: false,
		},
		{
			name:      "invalid line format (not enough parts)",
			line:      "2023-10-27 10:00 | Project A",
			want:      nil,
			wantErr:   true,
			errTarget: file.ErrSkip,
		},
		{
			name:    "invalid start time format",
			line:    "invalid-time | Project A | Description",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid end time format",
			line:    "2023-10-27 10:00 - invalid-time | Project A | Description",
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid activity with seconds",
			line: "2023-10-27 10:00:30 | Project C | Seconds test",
			want: &models.Activity{
				StartTime:   localTime(2023, 10, 27, 10, 0, 30),
				EndTime:     nil,
				Project:     "Project C",
				Description: "Seconds test",
				Attributes:  []models.Attribute{},
			},
			wantErr: false,
		},
		{
			name: "description with pipe character (quoted)",
			line: `abc123 | 2023-10-27 10:00 | Project A | "Working on feature X | Y | Z"`,
			want: &models.Activity{
				ID:          "abc123",
				StartTime:   localTime(2023, 10, 27, 10, 0, 0),
				EndTime:     nil,
				Project:     "Project A",
				Description: "Working on feature X | Y | Z",
				Attributes:  []models.Attribute{},
			},
			wantErr: false,
		},
		{
			name: "project with pipe character (quoted)",
			line: `def456 | 2023-10-27 11:00 | "Project A | B" | Meeting notes`,
			want: &models.Activity{
				ID:          "def456",
				StartTime:   localTime(2023, 10, 27, 11, 0, 0),
				EndTime:     nil,
				Project:     "Project A | B",
				Description: "Meeting notes",
				Attributes:  []models.Attribute{},
			},
			wantErr: false,
		},
		{
			name: "both project and description with pipes (quoted)",
			line: `ghi789 | 2023-10-27 12:00 - 2023-10-27 13:00 | "Project A | B" | "Task X | Y | Z"`,
			want: &models.Activity{
				ID:          "ghi789",
				StartTime:   localTime(2023, 10, 27, 12, 0, 0),
				EndTime:     func() *time.Time { t := localTime(2023, 10, 27, 13, 0, 0); return &t }(),
				Project:     "Project A | B",
				Description: "Task X | Y | Z",
				Attributes:  []models.Attribute{},
			},
			wantErr: false,
		},
		{
			name: "quoted field with escaped quote",
			line: `jkl012 | 2023-10-27 14:00 | Project D | "Description with \"quotes\" inside"`,
			want: &models.Activity{
				ID:          "jkl012",
				StartTime:   localTime(2023, 10, 27, 14, 0, 0),
				EndTime:     nil,
				Project:     "Project D",
				Description: `Description with "quotes" inside`,
				Attributes:  []models.Attribute{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := file.ParseActivity(tt.line)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errTarget != nil {
					require.ErrorIs(t, err, tt.errTarget)
				}
				return
			}
			if assert.NoError(t, err) {
				// For legacy format (no ID in input), ID is auto-generated, so we can't compare it directly
				if tt.want.ID == "" {
					assert.NotEmpty(t, got.ID, "ID should be auto-generated for legacy format")
					// Compare everything except ID
					assert.Equal(t, tt.want.StartTime.Unix(), got.StartTime.Unix())
					if tt.want.EndTime != nil {
						require.NotNil(t, got.EndTime)
						assert.Equal(t, tt.want.EndTime.Unix(), got.EndTime.Unix())
					} else {
						assert.Nil(t, got.EndTime)
					}
					assert.Equal(t, tt.want.Project, got.Project)
					assert.Equal(t, tt.want.Description, got.Description)
					assert.Equal(t, tt.want.Attributes, got.Attributes)
				} else {
					// For new format with explicit ID, compare everything
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.StartTime.Unix(), got.StartTime.Unix())
					if tt.want.EndTime != nil {
						require.NotNil(t, got.EndTime)
						assert.Equal(t, tt.want.EndTime.Unix(), got.EndTime.Unix())
					} else {
						assert.Nil(t, got.EndTime)
					}
					assert.Equal(t, tt.want.Project, got.Project)
					assert.Equal(t, tt.want.Description, got.Description)
					assert.Equal(t, tt.want.Attributes, got.Attributes)
				}
			}
		})
	}
}

func TestFormatActivity(t *testing.T) {
	localTime := func(year int, month time.Month, day, hour, minute, sec int) time.Time {
		return time.Date(year, month, day, hour, minute, sec, 0, time.Local)
	}

	tests := []struct {
		name     string
		activity models.Activity
		want     string
	}{
		{
			name: "activity with start and end time",
			activity: models.Activity{
				ID:          "abc123",
				StartTime:   localTime(2023, 10, 27, 10, 0, 0),
				EndTime:     func() *time.Time { t := localTime(2023, 10, 27, 11, 0, 0); return &t }(),
				Project:     "Project A",
				Description: "Description",
			},
			want: "abc123 | 2023-10-27 10:00:00 - 2023-10-27 11:00:00 | Project A | Description",
		},
		{
			name: "activity with start time only",
			activity: models.Activity{
				ID:          "def456",
				StartTime:   localTime(2023, 10, 27, 12, 0, 0),
				EndTime:     nil,
				Project:     "Project B",
				Description: "Meeting",
			},
			want: "def456 | 2023-10-27 12:00:00 | Project B | Meeting",
		},
		{
			name: "description with pipe character (should be quoted)",
			activity: models.Activity{
				ID:          "ghi789",
				StartTime:   localTime(2023, 10, 27, 10, 0, 0),
				EndTime:     nil,
				Project:     "Project A",
				Description: "Working on feature X | Y | Z",
			},
			want: `ghi789 | 2023-10-27 10:00:00 | Project A | "Working on feature X | Y | Z"`,
		},
		{
			name: "project with pipe character (should be quoted)",
			activity: models.Activity{
				ID:          "jkl012",
				StartTime:   localTime(2023, 10, 27, 11, 0, 0),
				EndTime:     nil,
				Project:     "Project A | B",
				Description: "Meeting notes",
			},
			want: `jkl012 | 2023-10-27 11:00:00 | "Project A | B" | Meeting notes`,
		},
		{
			name: "both project and description with pipes",
			activity: models.Activity{
				ID:          "mno345",
				StartTime:   localTime(2023, 10, 27, 12, 0, 0),
				EndTime:     func() *time.Time { t := localTime(2023, 10, 27, 13, 0, 0); return &t }(),
				Project:     "Project A | B",
				Description: "Task X | Y | Z",
			},
			want: `mno345 | 2023-10-27 12:00:00 - 2023-10-27 13:00:00 | "Project A | B" | "Task X | Y | Z"`,
		},
		{
			name: "description with quotes (should be escaped)",
			activity: models.Activity{
				ID:          "pqr678",
				StartTime:   localTime(2023, 10, 27, 14, 0, 0),
				EndTime:     nil,
				Project:     "Project D",
				Description: `Description with "quotes" inside`,
			},
			want: `pqr678 | 2023-10-27 14:00:00 | Project D | "Description with \"quotes\" inside"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := file.FormatActivity(tt.activity)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRoundTripParseFormat(t *testing.T) {
	localTime := func(year int, month time.Month, day, hour, minute, sec int) time.Time {
		return time.Date(year, month, day, hour, minute, sec, 0, time.Local)
	}

	tests := []struct {
		name     string
		activity models.Activity
	}{
		{
			name: "simple activity",
			activity: models.Activity{
				ID:          "abc123",
				StartTime:   localTime(2023, 10, 27, 10, 0, 0),
				EndTime:     func() *time.Time { t := localTime(2023, 10, 27, 11, 0, 0); return &t }(),
				Project:     "Project A",
				Description: "Simple description",
				Attributes:  []models.Attribute{},
			},
		},
		{
			name: "description with pipe",
			activity: models.Activity{
				ID:          "def456",
				StartTime:   localTime(2023, 10, 27, 12, 0, 0),
				EndTime:     nil,
				Project:     "Project B",
				Description: "Task A | Task B | Task C",
				Attributes:  []models.Attribute{},
			},
		},
		{
			name: "project with pipe",
			activity: models.Activity{
				ID:          "ghi789",
				StartTime:   localTime(2023, 10, 27, 13, 0, 0),
				EndTime:     nil,
				Project:     "Client A | Project X",
				Description: "Meeting",
				Attributes:  []models.Attribute{},
			},
		},
		{
			name: "both with pipes",
			activity: models.Activity{
				ID:          "jkl012",
				StartTime:   localTime(2023, 10, 27, 14, 0, 0),
				EndTime:     func() *time.Time { t := localTime(2023, 10, 27, 15, 0, 0); return &t }(),
				Project:     "Client A | Project Y",
				Description: "Feature A | Feature B",
				Attributes:  []models.Attribute{},
			},
		},
		{
			name: "description with quotes",
			activity: models.Activity{
				ID:          "mno345",
				StartTime:   localTime(2023, 10, 27, 16, 0, 0),
				EndTime:     nil,
				Project:     "Project C",
				Description: `Working on "important" feature`,
				Attributes:  []models.Attribute{},
			},
		},
		{
			name: "complex with pipes and quotes",
			activity: models.Activity{
				ID:          "pqr678",
				StartTime:   localTime(2023, 10, 27, 17, 0, 0),
				EndTime:     nil,
				Project:     `Client "ABC" | Project Z`,
				Description: `Task with | and "quotes"`,
				Attributes:  []models.Attribute{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Format the activity to a string
			formatted := file.FormatActivity(tt.activity)

			// Parse it back
			parsed, err := file.ParseActivity(formatted)
			require.NoError(t, err)
			require.NotNil(t, parsed)

			// Compare all fields
			assert.Equal(t, tt.activity.ID, parsed.ID)
			assert.Equal(t, tt.activity.StartTime.Unix(), parsed.StartTime.Unix())
			if tt.activity.EndTime != nil {
				require.NotNil(t, parsed.EndTime)
				assert.Equal(t, tt.activity.EndTime.Unix(), parsed.EndTime.Unix())
			} else {
				assert.Nil(t, parsed.EndTime)
			}
			assert.Equal(t, tt.activity.Project, parsed.Project)
			assert.Equal(t, tt.activity.Description, parsed.Description)
			assert.Equal(t, tt.activity.Attributes, parsed.Attributes)
		})
	}
}
