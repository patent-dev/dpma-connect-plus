package dpmaconnect

import (
	"encoding/xml"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// dpma-connect-plus speaks XML, so this is the XML analogue of the strict-decode
// (DisallowUnknownFields) JSON fixture tests used by the JSON clients in this
// repo. Every PARSED endpoint is asserted three ways over the committed
// testdata/*.xml fixtures (real recorded DPMA responses), in the normal build
// with no credentials:
//
//	(a) ELEMENT COVERAGE  (TestXMLCompleteness) - encoding/xml silently drops any
//	    element or attribute that no struct field maps to, so an incomplete struct
//	    loses real data without any error. For each fixture we collect every
//	    element path and attribute actually present in the XML, then assert the
//	    raw struct captures every one of them via an `xml:"..."` tag. A newly
//	    appearing (or previously dropped) element fails the test until it is
//	    modeled.
//	(b) GOLDEN ROUND-TRIP (TestXMLGoldenRoundTrip) - re-marshal the decoded raw
//	    struct and assert that, for every element path the struct models, the
//	    re-marshaled values reproduce the fixture's values. This proves the
//	    modeled projection is lossless: a value that decoded but did not survive a
//	    re-marshal (a wrong/ambiguous tag) is caught.
//	(c) KEY FIELDS        (TestParse* in xml_test.go) - 3-6 targeted value
//	    assertions per endpoint through the public Parse* functions, proving each
//	    value parsed into the right typed field.
//
// The six raw shapes here cover every parsed Client endpoint:
//	xmlPatentHitList        <- SearchPatentsParsed
//	xmlDPMAPatentDocument   <- GetPatentInfoParsed / GetPatentInfoByPublicationNumber
//	xmlTrademarkHitList     <- SearchTrademarksParsed
//	xmlTrademarkTransaction <- GetTrademarkInfoParsed
//	xmlDesignHitList        <- SearchDesignsParsed
//	xmlDesignTransaction    <- GetDesignInfoParsed
//
// RAW XML endpoints (Get...XML returning the raw bytes), PDF endpoints and
// image/thumbnail endpoints are NOT parsed, so they get a minimal well-formed /
// non-empty / format-magic contract instead of the three layers; those contracts
// are exercised live in integration_test.go (per-endpoint). None of the six raw
// structs uses a custom UnmarshalXML, so all six round-trip; if one ever did, its
// round-trip case would be skipped with a documented reason (see assertRoundTrip).
//
// The element/round-trip checks match encoding/xml semantics: matching is by
// local element name (namespace prefixes such as de: are ignored, exactly as the
// decoder ignores them), and paths are rooted at the response's top-level element.
//
// Round-trip NORMALIZATION rules (see roundTripValues / equalMultisetVals):
//   - whitespace: each captured chardata value is TrimSpace'd; values that are
//     empty after trimming are ignored (self-closing vs empty-element noise).
//   - namespace: prefixes are dropped; comparison is by local name only.
//   - order: values for a path are compared as a MULTISET (order-insensitive), so
//     sibling re-ordering by the marshaler does not cause a false diff.
//   - root: the leading root segment is stripped from every path so a raw struct
//     that re-marshals under a differing root name still lines up on its subtree.
//   - scope: only paths the struct MODELS are compared; unmodeled paths are the
//     job of layer (a) element coverage, not round-trip.

// schemaPaths reflects over a raw XML struct type and returns the set of
// element paths and attribute paths it can decode. Element paths look like
// "/Root/Child/Leaf"; attribute paths look like "/Root/Child @attrName".
func schemaPaths(t reflect.Type) (elements, attrs map[string]bool) {
	elements = map[string]bool{}
	attrs = map[string]bool{}
	seen := map[reflect.Type]bool{}

	// Determine the root element name from an XMLName field tag, if any.
	root := ""
	t = deref(t)
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Name == "XMLName" {
				if name := tagName(f.Tag.Get("xml")); name != "" {
					root = "/" + name
				}
			}
		}
	}
	elements[root] = true
	collectStruct(t, root, elements, attrs, seen)
	return elements, attrs
}

func collectStruct(t reflect.Type, prefix string, elements, attrs map[string]bool, seen map[reflect.Type]bool) {
	t = deref(t)
	if t.Kind() != reflect.Struct {
		return
	}
	// Guard against recursive types (none currently, but be safe).
	key := t
	if seen[key] {
		return
	}
	seen[key] = true
	defer delete(seen, key)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" { // unexported
			continue
		}
		tag := f.Tag.Get("xml")
		if f.Name == "XMLName" {
			continue
		}
		name, opts := parseTag(tag)
		if has(opts, "attr") {
			attrName := name
			if attrName == "" {
				attrName = f.Name
			}
			attrs[prefix+" @"+attrName] = true
			continue
		}
		// Non-element field kinds.
		if has(opts, "chardata") || has(opts, "innerxml") || has(opts, "comment") || has(opts, "cdata") {
			continue
		}
		if name == "-" {
			continue
		}
		if name == "" {
			name = f.Name
		}
		// Expand ">"-separated nested element paths.
		segs := strings.Split(name, ">")
		p := prefix
		for _, s := range segs {
			p += "/" + s
			elements[p] = true
		}
		collectStruct(f.Type, p, elements, attrs, seen)
	}
}

func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		t = t.Elem()
	}
	return t
}

func tagName(tag string) string {
	n, _ := parseTag(tag)
	return n
}

func parseTag(tag string) (name string, opts []string) {
	if tag == "" {
		return "", nil
	}
	parts := strings.Split(tag, ",")
	name = strings.TrimSpace(parts[0])
	// Strip any namespace prefix in the tag (e.g. "ns name") to local name.
	if idx := strings.LastIndex(name, " "); idx >= 0 {
		name = name[idx+1:]
	}
	return name, parts[1:]
}

func has(opts []string, want string) bool {
	for _, o := range opts {
		if o == want {
			return true
		}
	}
	return false
}

// localName returns the local part of an XML name (drops namespace).
func localName(name xml.Name) string {
	return name.Local
}

// fixturePaths walks the actual fixture XML and returns the element paths and
// attribute paths present, rooted and using local element names.
func fixturePaths(t *testing.T, data []byte) (elements, attrs map[string]bool) {
	t.Helper()
	elements = map[string]bool{}
	attrs = map[string]bool{}

	dec := xml.NewDecoder(strings.NewReader(string(data)))
	var stack []string
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch se := tok.(type) {
		case xml.StartElement:
			cur := strings.Join(stack, "") + "/" + localName(se.Name)
			stack = append(stack, "/"+localName(se.Name))
			elements[cur] = true
			for _, a := range se.Attr {
				// Skip namespace declarations.
				if a.Name.Space == "xmlns" || a.Name.Local == "xmlns" {
					continue
				}
				attrs[cur+" @"+a.Name.Local] = true
			}
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return elements, attrs
}

func assertComplete(t *testing.T, name string, data []byte, rawType reflect.Type) {
	t.Helper()
	schemaEls, schemaAttrs := schemaPaths(rawType)
	fixEls, fixAttrs := fixturePaths(t, data)

	var missingEls, missingAttrs []string
	for p := range fixEls {
		if !schemaEls[p] {
			missingEls = append(missingEls, p)
		}
	}
	for a := range fixAttrs {
		if !schemaAttrs[a] {
			missingAttrs = append(missingAttrs, a)
		}
	}
	sort.Strings(missingEls)
	sort.Strings(missingAttrs)

	if len(missingEls) > 0 {
		t.Errorf("%s: %d element path(s) present in the response are NOT captured by %s (data would be silently dropped):\n  %s",
			name, len(missingEls), rawType.Name(), strings.Join(missingEls, "\n  "))
	}
	if len(missingAttrs) > 0 {
		t.Errorf("%s: %d attribute(s) present in the response are NOT captured by %s:\n  %s",
			name, len(missingAttrs), rawType.Name(), strings.Join(missingAttrs, "\n  "))
	}
}

func TestXMLCompleteness(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		raw  any
	}{
		{"patent_search", patentSearchXML, xmlPatentHitList{}},
		{"patent_info", patentInfoXML, xmlDPMAPatentDocument{}},
		{"trademark_search", trademarkSearchXML, xmlTrademarkHitList{}},
		{"trademark_info", trademarkInfoXML, xmlTrademarkTransaction{}},
		{"design_search", designSearchXML, xmlDesignHitList{}},
		{"design_info", designInfoXML, xmlDesignTransaction{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertComplete(t, c.name, c.data, reflect.TypeOf(c.raw))
		})
	}
}

// roundTripCases is the authoritative table mapping each parsed shape to its
// fixture and raw unmarshal struct. It mirrors the TestXMLCompleteness table:
// the same six shapes cover every parsed Client endpoint.
func roundTripCases() []struct {
	name string
	data []byte
	raw  any
	skip string // non-empty: documented reason this shape cannot re-marshal
} {
	return []struct {
		name string
		data []byte
		raw  any
		skip string
	}{
		{name: "patent_search", data: patentSearchXML, raw: xmlPatentHitList{}},
		{name: "patent_info", data: patentInfoXML, raw: xmlDPMAPatentDocument{}},
		{name: "trademark_search", data: trademarkSearchXML, raw: xmlTrademarkHitList{}},
		{name: "trademark_info", data: trademarkInfoXML, raw: xmlTrademarkTransaction{}},
		{name: "design_search", data: designSearchXML, raw: xmlDesignHitList{}},
		{name: "design_info", data: designInfoXML, raw: xmlDesignTransaction{}},
	}
}

// TestXMLGoldenRoundTrip runs layer (b): decode each fixture into its raw struct,
// re-marshal, and assert that every element path the struct MODELS reproduces the
// fixture's values (whitespace-normalised, namespace-stripped, order-insensitive,
// root-relative). See the file header for the normalization rules.
func TestXMLGoldenRoundTrip(t *testing.T) {
	for _, c := range roundTripCases() {
		t.Run(c.name, func(t *testing.T) {
			if c.skip != "" {
				t.Skipf("round-trip not applicable: %s", c.skip)
			}
			assertRoundTrip(t, c.name, c.data, reflect.TypeOf(c.raw))
		})
	}
}

func assertRoundTrip(t *testing.T, name string, data []byte, rawType reflect.Type) {
	t.Helper()

	out := reflect.New(rawType).Interface()
	if err := xml.Unmarshal(data, out); err != nil {
		t.Fatalf("%s: unmarshal: %v", name, err)
	}
	remarshaled, err := xml.Marshal(out)
	if err != nil {
		t.Fatalf("%s: re-marshal: %v", name, err)
	}

	// Modeled element paths, made root-relative.
	schemaEls, _ := schemaPaths(rawType)
	schemaRel := stripRoots(schemaEls)

	fixVals := stripRootVals(roundTripValues(data))
	rtVals := stripRootVals(roundTripValues(remarshaled))

	var diffs []string
	for path, want := range fixVals {
		if !schemaRel[path] {
			continue // not modeled: governed by element coverage, not round-trip
		}
		got := rtVals[path]
		if !equalMultisetVals(want, got) {
			diffs = append(diffs, path+": fixture="+strings.Join(want, "|")+" remarshaled="+strings.Join(got, "|"))
		}
	}
	sort.Strings(diffs)
	if len(diffs) > 0 {
		if len(diffs) > 20 {
			diffs = append(diffs[:20], "...")
		}
		t.Errorf("%s: %d modeled element path(s) did not round-trip losslessly:\n  %s",
			name, len(diffs), strings.Join(diffs, "\n  "))
	}
}

// roundTripValues walks XML and returns, per rooted element path (local names),
// the multiset of non-empty trimmed character values directly under that element.
func roundTripValues(data []byte) map[string][]string {
	out := map[string][]string{}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	var stack []string
	var text strings.Builder
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch se := tok.(type) {
		case xml.StartElement:
			text.Reset()
			stack = append(stack, "/"+localName(se.Name))
		case xml.CharData:
			text.Write(se)
		case xml.EndElement:
			path := strings.Join(stack, "")
			if v := strings.TrimSpace(text.String()); v != "" {
				out[path] = append(out[path], v)
			}
			text.Reset()
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return out
}

// stripRoot drops the first "/segment" of a rooted path, leaving the path
// relative to the root element (e.g. "/a/b/c" -> "/b/c", "/a" -> "").
func stripRoot(p string) string {
	if len(p) == 0 || p[0] != '/' {
		return p
	}
	i := strings.IndexByte(p[1:], '/')
	if i < 0 {
		return ""
	}
	return p[1+i:]
}

func stripRoots(in map[string]bool) map[string]bool {
	out := make(map[string]bool, len(in))
	for p := range in {
		out[stripRoot(p)] = true
	}
	return out
}

func stripRootVals(in map[string][]string) map[string][]string {
	out := make(map[string][]string, len(in))
	for p, v := range in {
		rp := stripRoot(p)
		out[rp] = append(out[rp], v...)
	}
	return out
}

func equalMultisetVals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	ac := append([]string(nil), a...)
	bc := append([]string(nil), b...)
	sort.Strings(ac)
	sort.Strings(bc)
	for i := range ac {
		if ac[i] != bc[i] {
			return false
		}
	}
	return true
}
