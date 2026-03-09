package dpmaconnect

import (
	"testing"
	"time"
)

func TestValidatePatentQuery(t *testing.T) {
	if err := ValidatePatentQuery("TI=Elektrofahrzeug"); err != nil {
		t.Errorf("ValidatePatentQuery(valid) = %v", err)
	}
	if err := ValidatePatentQuery("MARKE=test"); err == nil {
		t.Error("ValidatePatentQuery(MARKE) = nil, want error")
	}
	if err := ValidatePatentQuery(""); err == nil {
		t.Error("ValidatePatentQuery(empty) = nil, want error")
	}
}

func TestValidateDesignQuery(t *testing.T) {
	if err := ValidateDesignQuery("INH=Samsung"); err != nil {
		t.Errorf("ValidateDesignQuery(valid) = %v", err)
	}
	if err := ValidateDesignQuery("IC=H02K"); err == nil {
		t.Error("ValidateDesignQuery(IC) = nil, want error")
	}
}

func TestValidateTrademarkQuery(t *testing.T) {
	if err := ValidateTrademarkQuery("md=Apple"); err != nil {
		t.Errorf("ValidateTrademarkQuery(valid) = %v", err)
	}
	if err := ValidateTrademarkQuery("TI=test"); err == nil {
		t.Error("ValidateTrademarkQuery(TI) = nil, want error")
	}
}

func TestFormatPublicationWeek(t *testing.T) {
	tests := []struct {
		name    string
		year    int
		week    int
		want    string
		wantErr bool
	}{
		{name: "regular week", year: 2024, week: 45, want: "202445"},
		{name: "single digit week", year: 2024, week: 5, want: "202405"},
		{name: "first week", year: 2024, week: 1, want: "202401"},
		{name: "last week", year: 2024, week: 53, want: "202453"},
		{name: "year 2000", year: 2000, week: 1, want: "200001"},
		{name: "week 0", year: 2024, week: 0, wantErr: true},
		{name: "week 54", year: 2024, week: 54, wantErr: true},
		{name: "negative week", year: 2024, week: -1, wantErr: true},
		{name: "year 0", year: 0, week: 1, wantErr: true},
		{name: "negative year", year: -1, week: 1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatPublicationWeek(tt.year, tt.week)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatPublicationWeek(%d, %d) error = %v, wantErr %v", tt.year, tt.week, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FormatPublicationWeek(%d, %d) = %s, want %s", tt.year, tt.week, got, tt.want)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{
			name: "regular date",
			date: time.Date(2024, 10, 23, 0, 0, 0, 0, time.UTC),
			want: "2024-10-23",
		},
		{
			name: "single digit month and day",
			date: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC),
			want: "2024-01-05",
		},
		{
			name: "first day of year",
			date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			want: "2024-01-01",
		},
		{
			name: "last day of year",
			date: time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
			want: "2024-12-31",
		},
		{
			name: "non-UTC timezone preserves local date",
			date: time.Date(2024, 10, 24, 1, 0, 0, 0, time.FixedZone("CET", 2*60*60)),
			want: "2024-10-24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDate(tt.date)
			if got != tt.want {
				t.Errorf("FormatDate() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestParsePublicationWeek(t *testing.T) {
	tests := []struct {
		name    string
		pubWeek string
		wantY   int
		wantW   int
		wantErr bool
	}{
		{
			name:    "valid week",
			pubWeek: "202445",
			wantY:   2024,
			wantW:   45,
			wantErr: false,
		},
		{
			name:    "valid first week",
			pubWeek: "202401",
			wantY:   2024,
			wantW:   1,
			wantErr: false,
		},
		{
			name:    "valid last week",
			pubWeek: "202453",
			wantY:   2024,
			wantW:   53,
			wantErr: false,
		},
		{
			name:    "too short",
			pubWeek: "20244",
			wantY:   0,
			wantW:   0,
			wantErr: true,
		},
		{
			name:    "too long",
			pubWeek: "2024455",
			wantY:   0,
			wantW:   0,
			wantErr: true,
		},
		{
			name:    "invalid week 0",
			pubWeek: "202400",
			wantY:   0,
			wantW:   0,
			wantErr: true,
		},
		{
			name:    "invalid week 54",
			pubWeek: "202454",
			wantY:   0,
			wantW:   0,
			wantErr: true,
		},
		{
			name:    "invalid year 0",
			pubWeek: "000045",
			wantY:   0,
			wantW:   0,
			wantErr: true,
		},
		{
			name:    "non-numeric",
			pubWeek: "20XX45",
			wantY:   0,
			wantW:   0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotY, gotW, err := ParsePublicationWeek(tt.pubWeek)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePublicationWeek(%s) error = %v, wantErr %v", tt.pubWeek, err, tt.wantErr)
				return
			}
			if gotY != tt.wantY || gotW != tt.wantW {
				t.Errorf("ParsePublicationWeek(%s) = (%d, %d), want (%d, %d)", tt.pubWeek, gotY, gotW, tt.wantY, tt.wantW)
			}
		})
	}
}

func TestValidatePatentQueryUnicode(t *testing.T) {
	if err := ValidatePatentQuery(`INH="München"`); err != nil {
		t.Errorf("ValidatePatentQuery with umlaut: %v", err)
	}
	if err := ValidatePatentQuery(`TI="Straßenführung"`); err != nil {
		t.Errorf("ValidatePatentQuery with ß: %v", err)
	}
	if err := ValidateDesignQuery(`INH="Müller"`); err != nil {
		t.Errorf("ValidateDesignQuery with umlaut: %v", err)
	}
}

func TestValidatePeriod(t *testing.T) {
	valid := []string{"daily", "weekly", "monthly", "yearly"}
	for _, p := range valid {
		if err := ValidatePeriod(p); err != nil {
			t.Errorf("ValidatePeriod(%q) = %v, want nil", p, err)
		}
	}

	invalid := []string{"", "banana", "Daily", "WEEKLY", "hourly"}
	for _, p := range invalid {
		if err := ValidatePeriod(p); err == nil {
			t.Errorf("ValidatePeriod(%q) = nil, want error", p)
		}
	}
}
