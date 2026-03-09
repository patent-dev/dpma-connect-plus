package query

import (
	"sort"
	"testing"
)

func TestParseQuery_ValidQueries(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		service Service
	}{
		{"simple field=value", "TI=Elektrofahrzeug", ServicePatent},
		{"AND query", "TI=Elektrofahrzeug AND INH=Siemens", ServicePatent},
		{"OR query", "TI=Motor OR TI=Antrieb", ServicePatent},
		{"German AND", "TI=Elektrofahrzeug UND INH=Siemens", ServicePatent},
		{"German OR", "TI=Motor ODER TI=Antrieb", ServicePatent},
		{"NOT query", "TI=Elektrofahrzeug NOT INH=Siemens", ServicePatent},
		{"German NOT", "TI=Elektrofahrzeug NICHT INH=Siemens", ServicePatent},
		{"parentheses", "(TI=Motor OR TI=Antrieb) AND IC=H02K?", ServicePatent},
		{"nested parens", "((TI=Motor) OR (TI=Antrieb)) AND INH=Siemens", ServicePatent},
		{"quoted value", "INH=\"Siemens AG\"", ServicePatent},
		{"date comparison >=", "AT>=01.01.2024", ServicePatent},
		{"date comparison <=", "PUB<=31.12.2024", ServicePatent},
		{"comparison >", "AT>01.01.2024", ServicePatent},
		{"comparison <", "PUB<31.12.2024", ServicePatent},
		{"IPC with wildcard", "IC=H02K?", ServicePatent},
		{"procedure data in braces", "{VST=pub-offenlegungschrift UND VSTT=05.01.2011}", ServicePatent},
		{"complex patent query", "SART=Patent AND PET=09.09.2010 AND IC=G01N?", ServicePatent},
		{"exists operator", "exists INH", ServicePatent},

		// Design queries
		{"design search", "INH=Samsung", ServiceDesign},
		{"design with class", "WKL=06-11 AND ERZ=Sportmatten", ServiceDesign},
		{"design active", "BA=aktiv AND (TI=\"Aufbewahrungsboxen\" OR ERZ=\"Aufbewahrungsboxen\") AND DB=DE", ServiceDesign},

		// Trademark queries
		{"trademark search", "MARKE=mars", ServiceTrademark},
		{"trademark md alias", "md=Apple", ServiceTrademark},
		{"trademark complex", "INH=\"München\" AND AT>=01.01.2010 AND DB=DE AND BA=eingetragen NOT KL=44", ServiceTrademark},
		{"trademark quoted", "MARKE=\"e-mail for you\"", ServiceTrademark},
		{"trademark wildcard prefix", "MARKE=?brain?", ServiceTrademark},
		{"trademark wildcard single", "MARKE=tele!om", ServiceTrademark},

		// ServiceAny accepts all fields
		{"any service", "TI=test AND MARKE=test", ServiceAny},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := ParseQuery(tt.query, tt.service)
			if err != nil {
				t.Fatalf("ParseQuery(%q) error = %v", tt.query, err)
			}
			if !q.Valid {
				t.Errorf("ParseQuery(%q) Valid = false, errors: %v", tt.query, q.Errors)
			}
			if q.Raw != tt.query {
				t.Errorf("ParseQuery(%q) Raw = %q", tt.query, q.Raw)
			}
		})
	}
}

func TestParseQuery_InvalidQueries(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		service Service
		wantErr string
	}{
		{"empty query", "", ServicePatent, "query cannot be empty"},
		{"whitespace only", "   ", ServicePatent, "query cannot be empty"},
		{"unmatched open paren", "(TI=test", ServicePatent, "unclosed parentheses"},
		{"unmatched close paren", "TI=test)", ServicePatent, "unmatched closing parenthesis"},
		{"unmatched open brace", "{VST=test", ServicePatent, "unclosed braces"},
		{"unmatched close brace", "VST=test}", ServicePatent, "unmatched closing brace"},
		{"no field=value pattern", "hello world", ServicePatent, "must contain at least one FIELD=value"},
		{"wrong field for service", "MARKE=test", ServicePatent, "unknown field"},
		{"wrong field for trademark", "IC=H02K", ServiceTrademark, "unknown field"},
		{"wrong field for design", "AB=abstract", ServiceDesign, "unknown field"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := ParseQuery(tt.query, tt.service)
			if tt.query == "" || tt.query == "   " {
				if err == nil {
					t.Errorf("ParseQuery(%q) expected error", tt.query)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseQuery(%q) unexpected error = %v", tt.query, err)
			}
			if q.Valid {
				t.Errorf("ParseQuery(%q) Valid = true, expected false", tt.query)
			}
			if q.Validate() == nil {
				t.Errorf("ParseQuery(%q) Validate() = nil, expected error", tt.query)
			}
			// Check that the error message contains the expected substring
			found := false
			for _, e := range q.Errors {
				if contains(e, tt.wantErr) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ParseQuery(%q) errors = %v, want error containing %q", tt.query, q.Errors, tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestParseQuery_Tokenization(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantFields []string
	}{
		{"single field", "TI=test", []string{"TI"}},
		{"two fields", "TI=test AND INH=Siemens", []string{"TI", "INH"}},
		{"three fields", "TI=test AND INH=Siemens OR IC=H02K", []string{"TI", "INH", "IC"}},
		{"duplicate fields", "TI=test OR TI=motor", []string{"TI"}},
		{"braces fields", "{VST=pub-offenlegungschrift UND VSTT=01.01.2011}", []string{"VST", "VSTT"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := ParseQuery(tt.query, ServicePatent)
			if err != nil {
				t.Fatalf("ParseQuery(%q) error = %v", tt.query, err)
			}

			fields := q.GetFields()
			sort.Strings(fields)
			sort.Strings(tt.wantFields)

			if len(fields) != len(tt.wantFields) {
				t.Errorf("GetFields() = %v, want %v", fields, tt.wantFields)
				return
			}
			for i := range fields {
				if fields[i] != tt.wantFields[i] {
					t.Errorf("GetFields() = %v, want %v", fields, tt.wantFields)
					break
				}
			}
		})
	}
}

func TestParseQuery_HasField(t *testing.T) {
	q, err := ParseQuery("TI=Elektrofahrzeug AND INH=Siemens", ServicePatent)
	if err != nil {
		t.Fatal(err)
	}

	if !q.HasField("TI") {
		t.Error("HasField(TI) = false, want true")
	}
	if !q.HasField("INH") {
		t.Error("HasField(INH) = false, want true")
	}
	if q.HasField("IC") {
		t.Error("HasField(IC) = true, want false")
	}
}

func TestParseQuery_URLEncode(t *testing.T) {
	tests := []struct {
		query string
		want  string
	}{
		{"TI=test", "TI%3Dtest"},
		{"TI=test AND INH=foo", "TI%3Dtest+AND+INH%3Dfoo"},
	}

	for _, tt := range tests {
		q, _ := ParseQuery(tt.query, ServiceAny)
		got := q.URLEncode()
		if got != tt.want {
			t.Errorf("URLEncode(%q) = %q, want %q", tt.query, got, tt.want)
		}
	}
}

func TestParseQuery_String(t *testing.T) {
	raw := "TI=Elektrofahrzeug AND INH=Siemens"
	q, _ := ParseQuery(raw, ServicePatent)
	if q.String() != raw {
		t.Errorf("String() = %q, want %q", q.String(), raw)
	}
}

func TestParseQuery_TokenCount(t *testing.T) {
	q, _ := ParseQuery("TI=test", ServicePatent)
	if q.TokenCount() < 3 {
		t.Errorf("TokenCount() = %d, want >= 3", q.TokenCount())
	}
}

func TestIsValidField(t *testing.T) {
	tests := []struct {
		field   string
		service Service
		want    bool
	}{
		// Patent fields
		{"TI", ServicePatent, true},
		{"INH", ServicePatent, true},
		{"IC", ServicePatent, true},
		{"AB", ServicePatent, true},
		{"CT", ServicePatent, true},
		{"SART", ServicePatent, true},
		{"MARKE", ServicePatent, false},
		{"ERZ", ServicePatent, false},

		// Design fields
		{"INH", ServiceDesign, true},
		{"ERZ", ServiceDesign, true},
		{"WKL", ServiceDesign, true},
		{"ENTW", ServiceDesign, true},
		{"IC", ServiceDesign, false},
		{"AB", ServiceDesign, false},

		// Trademark fields
		{"MARKE", ServiceTrademark, true},
		{"md", ServiceTrademark, true},
		{"INH", ServiceTrademark, true},
		{"KL", ServiceTrademark, true},
		{"BKL", ServiceTrademark, true},
		{"IC", ServiceTrademark, false},
		{"TI", ServiceTrademark, false},

		// ServiceAny
		{"TI", ServiceAny, true},
		{"MARKE", ServiceAny, true},
		{"ERZ", ServiceAny, true},
		{"NONEXISTENT", ServiceAny, false},
	}

	for _, tt := range tests {
		t.Run(tt.field+"_"+string(tt.service), func(t *testing.T) {
			got := IsValidField(tt.field, tt.service)
			if got != tt.want {
				t.Errorf("IsValidField(%q, %q) = %v, want %v", tt.field, tt.service, got, tt.want)
			}
		})
	}
}

func TestIsValidOperator(t *testing.T) {
	valid := []string{"AND", "OR", "NOT", "UND", "ODER", "NICHT", "and", "or", "not", "und", "oder", "nicht"}
	for _, op := range valid {
		if !IsValidOperator(op) {
			t.Errorf("IsValidOperator(%q) = false, want true", op)
		}
	}

	invalid := []string{"NAND", "XOR", "BUT", "WITH", ""}
	for _, op := range invalid {
		if IsValidOperator(op) {
			t.Errorf("IsValidOperator(%q) = true, want false", op)
		}
	}
}

func TestGetField(t *testing.T) {
	f, ok := GetField("TI", ServicePatent)
	if !ok {
		t.Fatal("GetField(TI, patent) not found")
	}
	if f.Code != "TI" {
		t.Errorf("field.Code = %q, want TI", f.Code)
	}
	if f.Description == "" {
		t.Error("field.Description is empty")
	}

	_, ok = GetField("NONEXISTENT", ServicePatent)
	if ok {
		t.Error("GetField(NONEXISTENT) found, want not found")
	}

	// ServiceAny finds fields from any service
	f, ok = GetField("MARKE", ServiceAny)
	if !ok {
		t.Error("GetField(MARKE, any) not found")
	}
	if f.Code != "MARKE" {
		t.Errorf("field.Code = %q, want MARKE", f.Code)
	}
}

func TestGetValidFields(t *testing.T) {
	patentFields := GetValidFields(ServicePatent)
	if len(patentFields) < 30 {
		t.Errorf("GetValidFields(patent) returned %d fields, want >= 30", len(patentFields))
	}

	designFields := GetValidFields(ServiceDesign)
	if len(designFields) < 10 {
		t.Errorf("GetValidFields(design) returned %d fields, want >= 10", len(designFields))
	}

	trademarkFields := GetValidFields(ServiceTrademark)
	if len(trademarkFields) < 10 {
		t.Errorf("GetValidFields(trademark) returned %d fields, want >= 10", len(trademarkFields))
	}

	allFields := GetValidFields(ServiceAny)
	if len(allFields) < len(patentFields) {
		t.Errorf("GetValidFields(any) returned %d fields, want >= %d", len(allFields), len(patentFields))
	}
}

func TestParseQuery_QuotedValues(t *testing.T) {
	q, err := ParseQuery("INH=\"Siemens AG\" AND TI=\"electric motor\"", ServicePatent)
	if err != nil {
		t.Fatal(err)
	}
	if !q.Valid {
		t.Errorf("Valid = false, errors: %v", q.Errors)
	}
	if !q.HasField("INH") || !q.HasField("TI") {
		t.Error("missing expected fields")
	}
}

func TestParseQuery_BraceQuery(t *testing.T) {
	q, err := ParseQuery("{VST=pub-offenlegungschrift UND VSTT=05.01.2011} UND IC=A47C15/00", ServicePatent)
	if err != nil {
		t.Fatal(err)
	}
	if !q.Valid {
		t.Errorf("Valid = false, errors: %v", q.Errors)
	}
	fields := q.GetFields()
	if len(fields) < 3 {
		t.Errorf("GetFields() = %v, want >= 3 fields", fields)
	}
}

func TestParseQuery_ComparisonOperators(t *testing.T) {
	queries := []string{
		"AT=01.01.2024",
		"AT>=01.01.2024",
		"AT<=31.12.2024",
		"AT>01.01.2024",
		"AT<31.12.2024",
	}
	for _, raw := range queries {
		q, err := ParseQuery(raw, ServicePatent)
		if err != nil {
			t.Fatalf("ParseQuery(%q) error = %v", raw, err)
		}
		if !q.Valid {
			t.Errorf("ParseQuery(%q) Valid = false, errors: %v", raw, q.Errors)
		}
		if !q.HasField("AT") {
			t.Errorf("ParseQuery(%q) HasField(AT) = false", raw)
		}
	}
}

func TestParseQuery_ComplexQueries(t *testing.T) {
	queries := []struct {
		name    string
		query   string
		service Service
	}{
		{
			"patent gazette search",
			"{VST=pub-offenlegungschrift UND VSTT=07.10.2010} UND IC=H05H?",
			ServicePatent,
		},
		{
			"SPC search",
			"SART=schutzzertifikat UND OT=21.10.2010 UND IC=C07?",
			ServicePatent,
		},
		{
			"EP patent search with procedure data",
			"{VST=ep-anmeldung-veroeffentlichung-ep-patenterteilung UND VSTT=30.06.2010} UND INH=maasland",
			ServicePatent,
		},
		{
			"trademark with multiple conditions",
			"INH=\"München\" AND AT>=01.01.2010 AND DB=DE AND BA=eingetragen NOT KL=44",
			ServiceTrademark,
		},
		{
			"design with active status",
			"BA=aktiv AND (TI=\"Aufbewahrungsboxen\" OR ERZ=\"Aufbewahrungsboxen\") AND DB=DE",
			ServiceDesign,
		},
	}

	for _, tt := range queries {
		t.Run(tt.name, func(t *testing.T) {
			q, err := ParseQuery(tt.query, tt.service)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if !q.Valid {
				t.Errorf("Valid = false, errors: %v", q.Errors)
			}
		})
	}
}
