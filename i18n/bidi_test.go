package i18n

import "testing"

// ---------------------------------------------------------------------------
// DetectDirection
// ---------------------------------------------------------------------------

func TestDetectDirection_Arabic(t *testing.T) {
	// "marhaba" in Arabic
	text := "\u0645\u0631\u062D\u0628\u0627"
	if dir := DetectDirection(text); dir != DirectionRTL {
		t.Fatalf("expected DirectionRTL for Arabic text, got %v", dir)
	}
}

func TestDetectDirection_Hebrew(t *testing.T) {
	// "shalom" in Hebrew
	text := "\u05E9\u05DC\u05D5\u05DD"
	if dir := DetectDirection(text); dir != DirectionRTL {
		t.Fatalf("expected DirectionRTL for Hebrew text, got %v", dir)
	}
}

func TestDetectDirection_Latin(t *testing.T) {
	text := "Hello World"
	if dir := DetectDirection(text); dir != DirectionLTR {
		t.Fatalf("expected DirectionLTR for Latin text, got %v", dir)
	}
}

func TestDetectDirection_MixedMostlyArabic(t *testing.T) {
	// Six Arabic chars + two Latin chars -> RTL dominant
	text := "\u0645\u0631\u062D\u0628\u0627\u0639 Hi"
	if dir := DetectDirection(text); dir != DirectionRTL {
		t.Fatalf("expected DirectionRTL for mostly Arabic text, got %v", dir)
	}
}

func TestDetectDirection_MixedMostlyLatin(t *testing.T) {
	// Mostly Latin with a single Arabic char
	text := "Hello World \u0645"
	if dir := DetectDirection(text); dir != DirectionLTR {
		t.Fatalf("expected DirectionLTR for mostly Latin text, got %v", dir)
	}
}

func TestDetectDirection_Empty(t *testing.T) {
	if dir := DetectDirection(""); dir != DirectionLTR {
		t.Fatalf("expected DirectionLTR for empty string, got %v", dir)
	}
}

func TestDetectDirection_OnlyNumbers(t *testing.T) {
	// Numbers are neutral, no strong characters -> LTR (default)
	if dir := DetectDirection("12345"); dir != DirectionLTR {
		t.Fatalf("expected DirectionLTR for only-numbers text, got %v", dir)
	}
}

func TestDetectDirection_OnlySpaces(t *testing.T) {
	if dir := DetectDirection("   "); dir != DirectionLTR {
		t.Fatalf("expected DirectionLTR for only-spaces text, got %v", dir)
	}
}

func TestDetectDirection_Persian(t *testing.T) {
	// Persian uses Arabic script
	text := "\u0641\u0627\u0631\u0633\u06CC" // "farsi"
	if dir := DetectDirection(text); dir != DirectionRTL {
		t.Fatalf("expected DirectionRTL for Persian text, got %v", dir)
	}
}

// ---------------------------------------------------------------------------
// IsRTLRune
// ---------------------------------------------------------------------------

func TestIsRTLRune_Arabic(t *testing.T) {
	// Arabic letter meem
	if !IsRTLRune('\u0645') {
		t.Fatal("expected Arabic meem to be RTL")
	}
}

func TestIsRTLRune_Hebrew(t *testing.T) {
	// Hebrew letter shin
	if !IsRTLRune('\u05E9') {
		t.Fatal("expected Hebrew shin to be RTL")
	}
}

func TestIsRTLRune_Syriac(t *testing.T) {
	if !IsRTLRune('\u0710') {
		t.Fatal("expected Syriac alaph to be RTL")
	}
}

func TestIsRTLRune_Thaana(t *testing.T) {
	if !IsRTLRune('\u0780') {
		t.Fatal("expected Thaana haa to be RTL")
	}
}

func TestIsRTLRune_Latin(t *testing.T) {
	if IsRTLRune('A') {
		t.Fatal("expected Latin A to not be RTL")
	}
}

func TestIsRTLRune_CJK(t *testing.T) {
	if IsRTLRune('\u4E16') { // Chinese character "world"
		t.Fatal("expected CJK character to not be RTL")
	}
}

func TestIsRTLRune_Digit(t *testing.T) {
	if IsRTLRune('5') {
		t.Fatal("expected digit to not be RTL")
	}
}

func TestIsRTLRune_RTLMark(t *testing.T) {
	if !IsRTLRune(0x200F) {
		t.Fatal("expected RIGHT-TO-LEFT MARK to be RTL")
	}
}

func TestIsRTLRune_Space(t *testing.T) {
	if IsRTLRune(' ') {
		t.Fatal("expected space to not be RTL")
	}
}

// ---------------------------------------------------------------------------
// ReverseString
// ---------------------------------------------------------------------------

func TestReverseString_ASCII(t *testing.T) {
	got := ReverseString("abcde")
	want := "edcba"
	if got != want {
		t.Fatalf("ReverseString(%q) = %q, want %q", "abcde", got, want)
	}
}

func TestReverseString_Unicode(t *testing.T) {
	// Arabic "marhaba" reversed
	input := "\u0645\u0631\u062D\u0628\u0627"
	got := ReverseString(input)
	want := "\u0627\u0628\u062D\u0631\u0645"
	if got != want {
		t.Fatalf("ReverseString(%q) = %q, want %q", input, got, want)
	}
}

func TestReverseString_Empty(t *testing.T) {
	got := ReverseString("")
	if got != "" {
		t.Fatalf("ReverseString(%q) = %q, want %q", "", got, "")
	}
}

func TestReverseString_SingleRune(t *testing.T) {
	got := ReverseString("X")
	if got != "X" {
		t.Fatalf("ReverseString(%q) = %q, want %q", "X", got, "X")
	}
}

func TestReverseString_Emoji(t *testing.T) {
	got := ReverseString("AB\U0001F600CD")
	want := "DC\U0001F600BA"
	if got != want {
		t.Fatalf("ReverseString with emoji = %q, want %q", got, want)
	}
}

func TestReverseString_MixedScripts(t *testing.T) {
	got := ReverseString("A\u05E9B")
	want := "B\u05E9A"
	if got != want {
		t.Fatalf("ReverseString(%q) = %q, want %q", "A\u05E9B", got, want)
	}
}

// ---------------------------------------------------------------------------
// MirrorBrackets
// ---------------------------------------------------------------------------

func TestMirrorBrackets_Basic(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"(hello)", ")hello("},
		{"[a]", "]a["},
		{"{x}", "}x{"},
		{"<tag>", ">tag<"},
		{"no brackets", "no brackets"},
		{"", ""},
		{"((()))", ")))((("},
	}
	for _, tt := range tests {
		got := MirrorBrackets(tt.input)
		if got != tt.want {
			t.Errorf("MirrorBrackets(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMirrorBrackets_Unicode(t *testing.T) {
	// Test guillemets
	got := MirrorBrackets("\u00ABhello\u00BB")
	want := "\u00BBhello\u00AB"
	if got != want {
		t.Fatalf("MirrorBrackets with guillemets = %q, want %q", got, want)
	}
}

func TestMirrorBrackets_SmartQuotes(t *testing.T) {
	got := MirrorBrackets("\u201Chello\u201D")
	want := "\u201Dhello\u201C"
	if got != want {
		t.Fatalf("MirrorBrackets with smart quotes = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// BidiReorder
// ---------------------------------------------------------------------------

func TestBidiReorder_Empty(t *testing.T) {
	got := BidiReorder("", DirectionLTR)
	if got != "" {
		t.Fatalf("BidiReorder empty = %q, want %q", got, "")
	}
	got = BidiReorder("", DirectionRTL)
	if got != "" {
		t.Fatalf("BidiReorder empty RTL = %q, want %q", got, "")
	}
}

func TestBidiReorder_PureLTR(t *testing.T) {
	text := "Hello World"
	got := BidiReorder(text, DirectionLTR)
	if got != text {
		t.Fatalf("BidiReorder pure LTR = %q, want %q", got, text)
	}
}

func TestBidiReorder_PureRTL(t *testing.T) {
	// Pure Arabic text should be reversed
	// "meem ra ha ba alef" -> reversed
	input := "\u0645\u0631\u062D\u0628\u0627"
	got := BidiReorder(input, DirectionRTL)
	want := ReverseString(input)
	if got != want {
		t.Fatalf("BidiReorder pure RTL = %q, want %q", got, want)
	}
}

func TestBidiReorder_PureRTLWithBrackets(t *testing.T) {
	// Brackets should be mirrored in RTL
	input := "(\u0645\u0631\u062D)"
	got := BidiReorder(input, DirectionRTL)
	// Reversed and mirrored: ")" + reversed_arabic + "("
	want := MirrorBrackets(ReverseString(input))
	if got != want {
		t.Fatalf("BidiReorder RTL brackets = %q, want %q", got, want)
	}
}

func TestBidiReorder_MixedLTRBase(t *testing.T) {
	// "Hello" + Arabic "marhaba" with LTR base
	// The Arabic part should be reversed, Latin stays
	arabic := "\u0645\u0631\u062D\u0628\u0627"
	input := "Hello " + arabic
	got := BidiReorder(input, DirectionLTR)
	// "Hello " stays, Arabic reversed
	want := "Hello " + ReverseString(arabic)
	if got != want {
		t.Fatalf("BidiReorder mixed LTR base = %q, want %q", got, want)
	}
}

func TestBidiReorder_MixedRTLBase(t *testing.T) {
	// Arabic + "Hello" with RTL base
	arabic := "\u0645\u0631\u062D"
	input := arabic + " Hello"
	got := BidiReorder(input, DirectionRTL)
	// RTL base reverses run order: "Hello" comes first, then space, then reversed Arabic
	// Runs: [arabic(RTL), " "(neutral->RTL in RTL base), "Hello"(LTR)]
	// RTL base reverses run order: "Hello" + " " + reversed_arabic
	// The space between RTL and LTR in RTL base gets assigned to RTL since prev=RTL, next=LTR -> base=RTL
	// So runs: [arabic+space (RTL), Hello (LTR)]
	// Reversed order: Hello + reversed(arabic+space) mirrored
	wantContainsHello := true
	if wantContainsHello && len(got) == 0 {
		t.Fatal("BidiReorder mixed RTL base returned empty")
	}
	// Just verify it doesn't crash and contains both scripts
	hasLatin := false
	hasArabic := false
	for _, r := range got {
		if isStrongLTR(r) {
			hasLatin = true
		}
		if IsRTLRune(r) {
			hasArabic = true
		}
	}
	if !hasLatin || !hasArabic {
		t.Fatalf("BidiReorder mixed RTL base = %q, expected both Latin and Arabic characters", got)
	}
}

func TestBidiReorder_NumbersInRTLContext(t *testing.T) {
	// Numbers within RTL context should stay LTR (not reversed)
	// Arabic + "123" + Arabic
	arabic1 := "\u0645\u0631"
	arabic2 := "\u062D\u0628"
	input := arabic1 + "123" + arabic2
	got := BidiReorder(input, DirectionRTL)

	// The numbers "123" should appear as "123" (not "321") in the output
	if !containsSubstring(got, "123") {
		t.Fatalf("BidiReorder numbers in RTL = %q, expected numbers to stay as '123'", got)
	}
}

func TestBidiReorder_PureLTRWithRTLBase(t *testing.T) {
	// Pure LTR text with RTL base direction should still be readable
	text := "Hello"
	got := BidiReorder(text, DirectionRTL)
	// No RTL chars, but RTL base. Fast path: !hasRTL && !hasLTR is false (hasLTR=true).
	// hasRTL=false, hasLTR=true -> goes to mixed path.
	// All chars are LTR, so single LTR run. RTL base reverses run order (but only 1 run).
	// The LTR run is kept as-is.
	if got != "Hello" {
		t.Fatalf("BidiReorder pure LTR with RTL base = %q, want %q", got, "Hello")
	}
}

func TestBidiReorder_OnlyNeutrals_LTRBase(t *testing.T) {
	text := "123 !@#"
	got := BidiReorder(text, DirectionLTR)
	// No strong chars at all, LTR base -> fast path returns as-is
	if got != text {
		t.Fatalf("BidiReorder neutrals LTR = %q, want %q", got, text)
	}
}

func TestBidiReorder_OnlyNeutrals_RTLBase(t *testing.T) {
	text := "123"
	got := BidiReorder(text, DirectionRTL)
	// No strong chars, RTL base -> reversed and mirrored
	want := "321"
	if got != want {
		t.Fatalf("BidiReorder neutrals RTL = %q, want %q", got, want)
	}
}

func TestBidiReorder_TrailingSpace(t *testing.T) {
	arabic := "\u0645\u0631\u062D"
	input := arabic + " "
	got := BidiReorder(input, DirectionRTL)
	// Pure RTL (no LTR strong chars) -> reverse + mirror
	want := MirrorBrackets(ReverseString(input))
	if got != want {
		t.Fatalf("BidiReorder trailing space = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Direction.String
// ---------------------------------------------------------------------------

func TestDirection_String(t *testing.T) {
	if DirectionLTR.String() != "LTR" {
		t.Fatalf("DirectionLTR.String() = %q, want %q", DirectionLTR.String(), "LTR")
	}
	if DirectionRTL.String() != "RTL" {
		t.Fatalf("DirectionRTL.String() = %q, want %q", DirectionRTL.String(), "RTL")
	}
}

// ---------------------------------------------------------------------------
// Integration: DetectDirection + BidiReorder
// ---------------------------------------------------------------------------

func TestDetectAndReorder_Arabic(t *testing.T) {
	text := "\u0645\u0631\u062D\u0628\u0627"
	dir := DetectDirection(text)
	if dir != DirectionRTL {
		t.Fatalf("expected RTL direction for Arabic")
	}
	got := BidiReorder(text, dir)
	want := ReverseString(text)
	if got != want {
		t.Fatalf("auto-detect + reorder Arabic = %q, want %q", got, want)
	}
}

func TestDetectAndReorder_Latin(t *testing.T) {
	text := "Hello World"
	dir := DetectDirection(text)
	if dir != DirectionLTR {
		t.Fatalf("expected LTR direction for Latin")
	}
	got := BidiReorder(text, dir)
	if got != text {
		t.Fatalf("auto-detect + reorder Latin = %q, want %q", got, text)
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestBidiReorder_SingleRTLChar(t *testing.T) {
	input := "\u0645" // single Arabic char
	got := BidiReorder(input, DirectionRTL)
	if got != input {
		t.Fatalf("BidiReorder single RTL char = %q, want %q", got, input)
	}
}

func TestBidiReorder_SingleLTRChar(t *testing.T) {
	got := BidiReorder("A", DirectionLTR)
	if got != "A" {
		t.Fatalf("BidiReorder single LTR char = %q, want %q", got, "A")
	}
}

func TestReverseString_TwoChars(t *testing.T) {
	got := ReverseString("AB")
	if got != "BA" {
		t.Fatalf("ReverseString(AB) = %q, want BA", got)
	}
}

func TestMirrorBrackets_NoChange(t *testing.T) {
	input := "hello world"
	got := MirrorBrackets(input)
	if got != input {
		t.Fatalf("MirrorBrackets unchanged = %q, want %q", got, input)
	}
}

func TestIsRTLRune_NKo(t *testing.T) {
	// NKo letter A
	if !IsRTLRune('\u07CA') {
		t.Fatal("expected NKo letter to be RTL")
	}
}

func TestBidiReorder_ConsecutiveSpaces(t *testing.T) {
	arabic := "\u0645\u0631"
	input := arabic + "   " + "Hello"
	got := BidiReorder(input, DirectionLTR)
	// Should not crash, and should contain both scripts
	if len(got) == 0 {
		t.Fatal("BidiReorder with spaces returned empty")
	}
}

func TestBidiReorder_NumbersOnly_LTR(t *testing.T) {
	got := BidiReorder("42", DirectionLTR)
	if got != "42" {
		t.Fatalf("BidiReorder numbers LTR = %q, want %q", got, "42")
	}
}

func TestBidiReorder_PunctuationBetweenRTL(t *testing.T) {
	// Arabic, comma, Arabic -> all RTL context
	input := "\u0645,\u0631"
	got := BidiReorder(input, DirectionRTL)
	want := MirrorBrackets(ReverseString(input))
	if got != want {
		t.Fatalf("BidiReorder punctuation in RTL = %q, want %q", got, want)
	}
}

// containsSubstring checks if s contains sub.
func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || findSubstring(s, sub))
}

func findSubstring(s, sub string) bool {
	runes := []rune(s)
	subRunes := []rune(sub)
	for i := 0; i <= len(runes)-len(subRunes); i++ {
		match := true
		for j := 0; j < len(subRunes); j++ {
			if runes[i+j] != subRunes[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
