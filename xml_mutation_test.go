package dpmaconnect

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// TestXMLMutationReachability is the mechanical "nothing is dropped" proof for
// the PUBLIC parsed types. TestXMLCompleteness asserts that the private raw
// structs model every element; this test goes one step further and asserts that
// every value survives the conversion into the exported Go value a caller
// actually receives.
//
// For every recorded response in testdata/ and for every non-blank text node
// and every attribute in it, the fixture bytes are changed at exactly that one
// place and parsed again through the public Parse* function. If the parsed
// value (ignoring RawXML, which trivially echoes the input) is identical to the
// unmutated parse, that node is unreachable through the typed API and the test
// fails naming the node.
//
// Not mutated, on purpose:
//   - the <?xml ...?> declaration and namespace declarations (xmlns, xmlns:*):
//     syntax of the envelope, not register data;
//   - fixtures the parser rejects as an error envelope (they yield an error
//     value, covered by errors_test.go).
func TestXMLMutationReachability(t *testing.T) {
	files, err := filepath.Glob("testdata/*.xml")
	if err != nil {
		t.Fatal(err)
	}
	parsers := map[string]func([]byte) (any, error){
		"PatentHitList":        func(b []byte) (any, error) { return ParsePatentSearch(b) },
		"dpma-patent-document": func(b []byte) (any, error) { return ParsePatentInfo(b) },
		"HitList":              func(b []byte) (any, error) { return ParseTrademarkSearch(b) },
		"DesignHitList":        func(b []byte) (any, error) { return ParseDesignSearch(b) },
	}
	covered := 0
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		root, firstChild := xmlRootAndBody(data)
		parse := parsers[root]
		if root == "Transaction" {
			switch firstChild {
			case "TradeMarkTransactionBody":
				parse = func(b []byte) (any, error) { return ParseTrademarkInfo(b) }
			case "DesignTransactionBody":
				parse = func(b []byte) (any, error) { return ParseDesignInfo(b) }
			}
		}
		if parse == nil {
			continue // bare <Error> envelope
		}
		base, err := parse(data)
		if err != nil {
			continue // error envelope
		}
		covered++
		t.Run(filepath.Base(file), func(t *testing.T) {
			nodes := xmlLeafNodes(t, data)
			if len(nodes) == 0 {
				t.Fatal("no text nodes or attributes found")
			}
			for _, n := range nodes {
				mutated := make([]byte, 0, len(data)+1)
				mutated = append(mutated, data[:n.start]...)
				mutated = append(mutated, mutateXMLValue(string(data[n.start:n.end]))...)
				mutated = append(mutated, data[n.end:]...)
				got, err := parse(mutated)
				if err != nil {
					continue // the value is consumed: changing it broke the typed parse
				}
				if reflect.DeepEqual(stripRawXML(base), stripRawXML(got)) {
					t.Errorf("unreachable: %s = %q is not represented in the parsed value", n.path, data[n.start:n.end])
				}
			}
		})
	}
	if covered < 15 {
		t.Fatalf("only %d fixtures were mutation-checked; testdata mapping is broken", covered)
	}
}

type xmlLeafNode struct {
	path       string
	start, end int // byte range of the raw value inside the fixture
}

// xmlRootAndBody returns the local name of the root element and of its first
// child element that ends in "TransactionBody" (to tell ST66 from ST86).
func xmlRootAndBody(data []byte) (root, body string) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return root, body
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if root == "" {
			root = se.Name.Local
			continue
		}
		if strings.HasSuffix(se.Name.Local, "TransactionBody") {
			return root, se.Name.Local
		}
	}
}

var xmlAttrRe = regexp.MustCompile(`([A-Za-z_][\w.:-]*)\s*=\s*(?:"([^"]*)"|'([^']*)')`)

// xmlLeafNodes lists every non-blank text node and every attribute of the
// document with the byte range of its raw (still escaped) value.
func xmlLeafNodes(t *testing.T, data []byte) []xmlLeafNode {
	t.Helper()
	var nodes []xmlLeafNode
	var stack []string
	seen := map[string]int{}
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		start := int(dec.InputOffset())
		tok, err := dec.RawToken()
		if err == io.EOF {
			return nodes
		}
		if err != nil {
			t.Fatalf("tokenize: %v", err)
		}
		end := int(dec.InputOffset())
		switch tk := tok.(type) {
		case xml.StartElement:
			name := tk.Name.Local
			if tk.Name.Space != "" {
				name = tk.Name.Space + ":" + name
			}
			stack = append(stack, name)
			path := "/" + strings.Join(stack, "/")
			seen[path]++
			path = fmt.Sprintf("%s[%d]", path, seen[path])
			for _, m := range xmlAttrRe.FindAllSubmatchIndex(data[start:end], -1) {
				attr := string(data[start+m[2] : start+m[3]])
				if attr == "xmlns" || strings.HasPrefix(attr, "xmlns:") {
					continue
				}
				vs, ve := m[4], m[5]
				if vs < 0 {
					vs, ve = m[6], m[7]
				}
				nodes = append(nodes, xmlLeafNode{path: path + " @" + attr, start: start + vs, end: start + ve})
			}
		case xml.EndElement:
			// A self-closing element yields a synthetic EndElement too.
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			raw := data[start:end]
			if len(bytes.TrimSpace(raw)) == 0 || len(stack) == 0 || bytes.HasPrefix(raw, []byte("<![CDATA[")) {
				continue
			}
			// Mutate inside the surrounding whitespace so trimming cannot hide it.
			lead := len(raw) - len(bytes.TrimLeft(raw, " \t\r\n"))
			trail := len(raw) - len(bytes.TrimRight(raw, " \t\r\n"))
			path := "/" + strings.Join(stack, "/")
			nodes = append(nodes, xmlLeafNode{path: fmt.Sprintf("%s[%d] text", path, seen[path]), start: start + lead, end: end - trail})
		}
	}
}

// mutateXMLValue returns a different value of the same lexical kind, so typed
// fields (counts, booleans) still parse and the change must show in the result.
func mutateXMLValue(v string) string {
	switch v {
	case "true":
		return "false"
	case "false":
		return "true"
	case "":
		return "x"
	}
	last := v[len(v)-1]
	if last >= '0' && last <= '9' {
		return v[:len(v)-1] + string('0'+(last-'0'+1)%10)
	}
	if last == 'x' {
		return v[:len(v)-1] + "y"
	}
	return v[:len(v)-1] + "x"
}

// stripRawXML returns a copy of a parsed result with its RawXML field cleared.
func stripRawXML(v any) any {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return v
	}
	cp := reflect.New(rv.Elem().Type())
	cp.Elem().Set(rv.Elem())
	if f := cp.Elem().FieldByName("RawXML"); f.IsValid() && f.CanSet() {
		f.Set(reflect.Zero(f.Type()))
	}
	return cp.Interface()
}
