package linthost

import "testing"

// TestUnicornStringContentPreservesControlEscapesInLiteralFix verifies
// control characters in the cooked value are re-spelled the way upstream's
// jsesc-based escapeString helper spells them.
//
// A literal like `"no\n"` cooks to a real newline; the fixer must not write
// that byte raw into the source (it would split the literal across lines).
// jsesc's minimal mode uses the named escapes for `\b \f \n \r \t`, spells a
// NUL as `\0` only when no decimal digit follows (a digit-adjacent NUL stays
// a raw byte because `\0` + digit would parse as a legacy octal escape), and
// passes every other control character — `\v`, C0 controls, DEL — through
// raw. Each expectation below was generated from the real
// `jsesc(value, {quotes, wrap: true, es6: true, minimal: true})` call.
//
//  1. Configure `{no: "yes"}` and fix literals containing named-escape
//     controls, NUL before end / letter / digit, and raw-pass controls.
//  2. Compare each rewritten literal with the upstream escape spelling.
//  3. Re-parse the fixed source, assert it stays parse-valid, and assert the
//     canonical output no longer fires.
func TestUnicornStringContentPreservesControlEscapesInLiteralFix(t *testing.T) {
  options := `{"patterns":{"no":"yes"}}`
  cases := []struct {
    name     string
    source   string
    expected string
  }{
    {
      name:     "newline escape",
      source:   `const foo = "no\n";` + "\n",
      expected: `const foo = "yes\n";` + "\n",
    },
    {
      name:     "carriage return escape",
      source:   `const foo = "no\r";` + "\n",
      expected: `const foo = "yes\r";` + "\n",
    },
    {
      name:     "tab escape",
      source:   `const foo = "no\t";` + "\n",
      expected: `const foo = "yes\t";` + "\n",
    },
    {
      name:     "backspace and form feed escapes",
      source:   `const foo = "no\b\f";` + "\n",
      expected: `const foo = "yes\b\f";` + "\n",
    },
    {
      name:     "null byte before end",
      source:   `const foo = "no\0";` + "\n",
      expected: `const foo = "yes\0";` + "\n",
    },
    {
      name:     "null byte before letter",
      source:   `const foo = "no\0a";` + "\n",
      expected: `const foo = "yes\0a";` + "\n",
    },
    {
      name:     "null byte before digit stays raw",
      source:   `const foo = "no\x001";` + "\n",
      expected: "const foo = \"yes\x001\";\n",
    },
    {
      name:     "vertical tab passes through raw",
      source:   `const foo = "no\v";` + "\n",
      expected: "const foo = \"yes\x0b\";\n",
    },
    {
      name:     "C0 control and DEL pass through raw",
      source:   `const foo = "no\x01\x7f";` + "\n",
      expected: "const foo = \"yes\x01\x7f\";\n",
    },
  }
  for _, test := range cases {
    t.Run(test.name, func(t *testing.T) {
      assertFixSnapshotWithOptions(t, "unicorn/string-content", test.source, options, test.expected)
      file := parseTSFile(t, "/virtual/fixed-string-content-controls.ts", test.expected)
      if diagnostics := file.Diagnostics(); len(diagnostics) != 0 {
        t.Fatalf("fixed source has parse diagnostics: %+v\n%s", diagnostics, test.expected)
      }
      assertRuleSkipsSourceWithOptions(t, "unicorn/string-content", test.expected, options)
    })
  }
}
