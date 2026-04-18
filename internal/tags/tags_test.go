package tags

import "testing"

var cases = []struct {
	name string
	in   string
	want string
}{
	{"no tags", "Hello world", "Hello world"},
	{"italic", "<i>word</i>", "_word_"},
	{"bold stripped", "<b>word</b>", "word"},
	{"underline stripped", "<u>word</u>", "word"},
	{"font stripped", `<font color="red">text</font>`, "text"},
	{"italic uppercase", "<I>word</I>", "_word_"},
	{"ssa an8", `{\an8}Hello`, "Hello"},
	{"ssa pos", `{\pos(10,20)}Hello`, "Hello"},
	{"nested b+i", "<b><i>text</i></b>", "_text_"},
	{"italic mid-sentence", "Say <i>this</i> now", "Say _this_ now"},
	{"multiple tags", "<b>bold</b> and <i>italic</i>", "bold and _italic_"},
	{"empty tag content", "<b></b>", ""},
	{"ssa + html", `{\an8}<i>styled</i>`, "_styled_"},
	{"plain passthrough", "just text, no tags!", "just text, no tags!"},
	{"unicode", "<i>日本語</i>", "_日本語_"},
}

func TestProcess(t *testing.T) {
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Process(c.in)
			if got != c.want {
				t.Errorf("Process(%q)\n got  %q\n want %q", c.in, got, c.want)
			}
		})
	}
}
