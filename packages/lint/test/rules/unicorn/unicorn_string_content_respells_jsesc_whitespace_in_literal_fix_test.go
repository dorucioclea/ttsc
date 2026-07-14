package linthost

import "testing"

// TestUnicornStringContentRespellsJsescWhitespaceInLiteralFix verifies the
// literal fixer escapes jsesc's exotic-whitespace class and nothing else.
//
// jsesc's minimal mode still escapes its `regexWhitespace` characters — NBSP,
// U+1680, U+2000–U+200A, U+2028, U+2029, U+202F, U+205F, U+3000 — as
// uppercase-hex `\xXX` / `\uXXXX`, while ordinary non-ASCII text and astral
// symbols pass through raw. U+2028/U+2029 are the load-bearing half (before
// ES2019 they were line terminators, and jsesc still refuses to write them
// raw); the exact spelling (two-digit `\xA0` for NBSP, uppercase hex digits)
// is pinned against the real jsesc output. Escape spellings inside the
// sources below are assembled by concatenation so every byte is explicit.
//
//  1. Configure `{no: "yes"}` and fix literals whose cooked values carry
//     exotic whitespace, plain accented text, and an astral emoji.
//  2. Compare each rewritten literal with the upstream jsesc spelling.
//  3. Re-parse the fixed source, assert it stays parse-valid, and assert the
//     canonical output no longer fires.
func TestUnicornStringContentRespellsJsescWhitespaceInLiteralFix(t *testing.T) {
  const backslashU = "\\u"
  options := `{"patterns":{"no":"yes"}}`
  cases := []struct {
    name     string
    source   string
    expected string
  }{
    {
      name: "escaped no-break space uses two-digit hex",
      // The four-digit source escape cooks to NBSP; jsesc re-spells it
      // with the shorter `\xA0` form.
      source:   `const foo = "no` + backslashU + `00A0after";` + "\n",
      expected: `const foo = "yes\xA0after";` + "\n",
    },
    {
      name: "raw no-break space is escaped too",
      // A literal NBSP byte in the source cooks to the same character and
      // must come back escaped, not raw.
      source:   "const foo = \"no after\";\n",
      expected: `const foo = "yes\xA0after";` + "\n",
    },
    {
      name: "line and paragraph separators use four-digit hex",
      source: `const foo = "no` + backslashU + `2028no` + backslashU + `2029";` + "\n",
      expected: `const foo = "yes` + backslashU + `2028yes` + backslashU + `2029";` + "\n",
    },
    {
      name: "hair and narrow spaces keep uppercase hex digits",
      source: `const foo = "no` + backslashU + `200ano` + backslashU + `202f";` + "\n",
      expected: `const foo = "yes` + backslashU + `200Ayes` + backslashU + `202F";` + "\n",
    },
    {
      name:     "accented text and astral symbols stay raw",
      source:   `const foo = "no héllo 🦄";` + "\n",
      expected: `const foo = "yes héllo 🦄";` + "\n",
    },
  }
  for _, test := range cases {
    t.Run(test.name, func(t *testing.T) {
      assertFixSnapshotWithOptions(t, "unicorn/string-content", test.source, options, test.expected)
      file := parseTSFile(t, "/virtual/fixed-string-content-whitespace.ts", test.expected)
      if diagnostics := file.Diagnostics(); len(diagnostics) != 0 {
        t.Fatalf("fixed source has parse diagnostics: %+v\n%s", diagnostics, test.expected)
      }
      assertRuleSkipsSourceWithOptions(t, "unicorn/string-content", test.expected, options)
    })
  }
}
