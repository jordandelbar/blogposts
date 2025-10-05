package validation

import (
	"reflect"
	"testing"
)

type mockFieldLevel struct {
	field reflect.Value
}

func (m *mockFieldLevel) Top() reflect.Value      { return reflect.Value{} }
func (m *mockFieldLevel) Parent() reflect.Value   { return reflect.Value{} }
func (m *mockFieldLevel) Field() reflect.Value    { return m.field }
func (m *mockFieldLevel) FieldName() string       { return "" }
func (m *mockFieldLevel) StructFieldName() string { return "" }
func (m *mockFieldLevel) Param() string           { return "" }
func (m *mockFieldLevel) GetTag() string          { return "" }
func (m *mockFieldLevel) ExtractType(field reflect.Value) (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, reflect.Invalid, false
}
func (m *mockFieldLevel) GetStructFieldOK() (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, reflect.Invalid, false
}
func (m *mockFieldLevel) GetStructFieldOKAdvanced(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool) {
	return reflect.Value{}, reflect.Invalid, false
}
func (m *mockFieldLevel) GetStructFieldOK2() (reflect.Value, reflect.Kind, bool, bool) {
	return reflect.Value{}, reflect.Invalid, false, false
}
func (m *mockFieldLevel) GetStructFieldOKAdvanced2(val reflect.Value, namespace string) (reflect.Value, reflect.Kind, bool, bool) {
	return reflect.Value{}, reflect.Invalid, false, false
}

func TestValidateNoHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid inputs (should return true)
		{
			name:     "plain text",
			input:    "This is plain text",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "alphanumeric with spaces",
			input:    "Hello World 123",
			expected: true,
		},
		{
			name:     "text with special characters",
			input:    "Text with !@#$%^&*()_+-=[]{}|;':\",./?",
			expected: true,
		},
		{
			name:     "unicode characters",
			input:    "Hello 世界 🌍",
			expected: true,
		},
		{
			name:     "newlines and tabs",
			input:    "Text with\nnewlines\tand\ttabs",
			expected: true,
		},
		{
			name:     "mathematical expressions",
			input:    "5 + 3 = 8 and 10 - 2 = 8",
			expected: true,
		},
		{
			name:     "HTML entities attempt",
			input:    "Text with &lt; escaped &gt;",
			expected: true, // HTML entities don't contain < or > directly
		},

		// Invalid inputs (should return false - contain HTML characters)
		{
			name:     "contains less than",
			input:    "This text contains < character",
			expected: false,
		},
		{
			name:     "contains greater than",
			input:    "This text contains > character",
			expected: false,
		},
		{
			name:     "contains both brackets",
			input:    "This text contains < and > characters",
			expected: false,
		},
		{
			name:     "HTML tag",
			input:    "<p>This is HTML</p>",
			expected: false,
		},
		{
			name:     "script tag",
			input:    "<script>alert('xss')</script>",
			expected: false,
		},
		{
			name:     "incomplete HTML tag",
			input:    "Text with < incomplete tag",
			expected: false,
		},
		{
			name:     "angle brackets in middle",
			input:    "Before < middle > after",
			expected: false,
		},
		{
			name:     "only less than",
			input:    "<",
			expected: false,
		},
		{
			name:     "only greater than",
			input:    ">",
			expected: false,
		},
		{
			name:     "mixed content with HTML",
			input:    "Normal text <b>bold</b> more text",
			expected: false,
		},
		{
			name:     "comparison operators",
			input:    "a < b and c > d",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := struct {
				Field string `validate:"no_html"`
			}{
				Field: tt.input,
			}

			structValue := reflect.ValueOf(testStruct)
			field := structValue.Field(0)

			mockFieldLevel := &mockFieldLevel{
				field: field,
			}

			result := validateNoHTML(mockFieldLevel)
			if result != tt.expected {
				t.Errorf("validateNoHTML(%q) = %v; expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateNoScriptTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid inputs (should return true)
		{
			name:     "plain text",
			input:    "This is plain text",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "HTML without script tags",
			input:    "<p>This is a paragraph</p>",
			expected: true,
		},
		{
			name:     "word script in normal context",
			input:    "I need to write a script for the movie",
			expected: true,
		},
		{
			name:     "word javascript in normal context",
			input:    "I'm learning JavaScript programming",
			expected: true,
		},
		{
			name:     "HTML with other tags",
			input:    "<div><span>Content</span></div>",
			expected: true,
		},
		{
			name:     "CSS styles",
			input:    "body { color: red; }",
			expected: true,
		},
		{
			name:     "URL with script in path",
			input:    "https://example.com/script/file.txt",
			expected: true,
		},
		{
			name:     "text with script word",
			input:    "The script was well written",
			expected: true,
		},
		{
			name:     "script tag with spaces",
			input:    "< script >alert('xss')</script>",
			expected: true, // This should pass because there's a space after <
		},

		// Invalid inputs (should return false - contain script-related content)
		{
			name:     "script tag lowercase",
			input:    "<script>alert('xss')</script>",
			expected: false,
		},
		{
			name:     "script tag uppercase",
			input:    "<SCRIPT>alert('xss')</SCRIPT>",
			expected: false,
		},
		{
			name:     "script tag mixed case",
			input:    "<ScRiPt>alert('xss')</ScRiPt>",
			expected: false,
		},
		{
			name:     "script tag with attributes",
			input:    "<script type='text/javascript'>code</script>",
			expected: false,
		},
		{
			name:     "script tag with src",
			input:    "<script src='malicious.js'></script>",
			expected: false,
		},
		{
			name:     "incomplete script tag",
			input:    "<script>incomplete",
			expected: false,
		},
		{
			name:     "javascript protocol lowercase",
			input:    "javascript:alert('xss')",
			expected: false,
		},
		{
			name:     "javascript protocol uppercase",
			input:    "JAVASCRIPT:alert('xss')",
			expected: false,
		},
		{
			name:     "javascript protocol mixed case",
			input:    "JavaScript:alert('xss')",
			expected: false,
		},
		{
			name:     "javascript in href attribute",
			input:    "<a href='javascript:void(0)'>Link</a>",
			expected: false,
		},
		{
			name:     "mixed script and javascript",
			input:    "<script>window.location='javascript:alert(1)'</script>",
			expected: false,
		},
		{
			name:     "script tag in middle of text",
			input:    "Before <script>alert('xss')</script> after",
			expected: false,
		},
		{
			name:     "javascript protocol in middle",
			input:    "Click here: javascript:alert('xss') for popup",
			expected: false,
		},
		{
			name:     "multiple script tags",
			input:    "<script>first</script><script>second</script>",
			expected: false,
		},
		{
			name:     "script tag with newlines",
			input:    "<script>\nalert('xss')\n</script>",
			expected: false,
		},
		{
			name:     "self-closing script tag",
			input:    "<script/>",
			expected: false,
		},
		{
			name:     "script tag with CDATA",
			input:    "<script><![CDATA[alert('xss')]]></script>",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := struct {
				Field string `validate:"no_script_tags"`
			}{
				Field: tt.input,
			}

			structValue := reflect.ValueOf(testStruct)
			field := structValue.Field(0)

			mockFieldLevel := &mockFieldLevel{
				field: field,
			}

			result := validateNoScriptTags(mockFieldLevel)
			if result != tt.expected {
				t.Errorf("validateNoScriptTags(%q) = %v; expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateAlphanumHyphen(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid inputs (should return true)
		{
			name:     "simple alphanumeric",
			input:    "hello",
			expected: true,
		},
		{
			name:     "alphanumeric with numbers",
			input:    "hello123",
			expected: true,
		},
		{
			name:     "numbers only",
			input:    "12345",
			expected: true,
		},
		{
			name:     "letters only uppercase",
			input:    "HELLO",
			expected: true,
		},
		{
			name:     "mixed case letters",
			input:    "HelloWorld",
			expected: true,
		},
		{
			name:     "single hyphen in middle",
			input:    "hello-world",
			expected: true,
		},
		{
			name:     "multiple hyphens in middle",
			input:    "hello-world-test",
			expected: true,
		},
		{
			name:     "alphanumeric with hyphens",
			input:    "hello-123-world",
			expected: true,
		},
		{
			name:     "single character",
			input:    "a",
			expected: true,
		},
		{
			name:     "single number",
			input:    "1",
			expected: true,
		},
		{
			name:     "long valid slug",
			input:    "this-is-a-very-long-slug-with-many-words-123",
			expected: true,
		},
		{
			name:     "consecutive hyphens in middle",
			input:    "hello--world",
			expected: true,
		},

		// Invalid inputs (should return false)
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "starts with hyphen",
			input:    "-hello",
			expected: false,
		},
		{
			name:     "ends with hyphen",
			input:    "hello-",
			expected: false,
		},
		{
			name:     "starts and ends with hyphen",
			input:    "-hello-",
			expected: false,
		},
		{
			name:     "only hyphens",
			input:    "---",
			expected: false,
		},
		{
			name:     "single hyphen",
			input:    "-",
			expected: false,
		},
		{
			name:     "contains spaces",
			input:    "hello world",
			expected: false,
		},
		{
			name:     "contains underscore",
			input:    "hello_world",
			expected: false,
		},
		{
			name:     "contains period",
			input:    "hello.world",
			expected: false,
		},
		{
			name:     "contains special characters",
			input:    "hello@world",
			expected: false,
		},
		{
			name:     "unicode characters",
			input:    "hello世界",
			expected: false,
		},
		{
			name:     "emoji",
			input:    "hello🌍",
			expected: false,
		},
		{
			name:     "accented characters",
			input:    "héllo",
			expected: false,
		},
		{
			name:     "newline character",
			input:    "hello\nworld",
			expected: false,
		},
		{
			name:     "tab character",
			input:    "hello\tworld",
			expected: false,
		},
		{
			name:     "leading space",
			input:    " hello",
			expected: false,
		},
		{
			name:     "trailing space",
			input:    "hello ",
			expected: false,
		},
		{
			name:     "multiple leading hyphens",
			input:    "--hello",
			expected: false,
		},
		{
			name:     "multiple trailing hyphens",
			input:    "hello--",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := struct {
				Field string `validate:"alphanum_hyphen"`
			}{
				Field: tt.input,
			}

			structValue := reflect.ValueOf(testStruct)
			field := structValue.Field(0)

			mockFieldLevel := &mockFieldLevel{
				field: field,
			}

			result := validateAlphanumHyphen(mockFieldLevel)
			if result != tt.expected {
				t.Errorf("validateAlphanumHyphen(%q) = %v; expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
