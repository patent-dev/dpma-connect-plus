package dpmaconnect

import (
	"encoding/xml"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// This is the XML analogue of a strict-decode (DisallowUnknownFields) JSON test.
//
// encoding/xml silently drops any element or attribute that no struct field
// maps to, so an incomplete struct loses real data without any error. These
// tests guard against that regression: for each recorded response fixture they
// collect every element path and attribute actually present in the XML, then
// assert that the corresponding raw struct captures every one of them via an
// `xml:"..."` tag. A newly appearing (or previously dropped) element makes the
// test fail until it is modeled.
//
// The check matches encoding/xml semantics: matching is by local element name
// (namespace prefixes such as de: are ignored, exactly as the decoder ignores
// them), and paths are rooted at the response's top-level element.

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
