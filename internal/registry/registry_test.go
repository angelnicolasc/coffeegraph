package registry

import "testing"

func TestParseListFromText(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{name: "empty", in: "", want: 0},
		{name: "single", in: "sales", want: 1},
		{name: "multi", in: "a\n\nb\nc", want: 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseListFromText(tc.in)
			if len(got) != tc.want {
				t.Fatalf("len(ParseListFromText(%q)) = %d, want %d", tc.in, len(got), tc.want)
			}
		})
	}
}

func TestPublishTemplateURL(t *testing.T) {
	u := PublishTemplateURL([]string{"one"})
	if u == "" {
		t.Fatalf("PublishTemplateURL() returned empty")
	}
}
