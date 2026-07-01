package telegram

import "testing"

func TestTagCaption(t *testing.T) {
	cases := []struct{ folder, name, want string }{
		{"", "photo.png", "photo.png"},
		{"root", "doc.pdf", "doc.pdf"},
		{"Documents", "file.txt", "#dir:Documents/file.txt"},
		{"Projects/2024", "data.csv", "#dir:Projects/2024/data.csv"},
	}
	for _, c := range cases {
		got := tagCaption(c.folder, c.name)
		if got != c.want {
			t.Errorf("tagCaption(%q,%q) = %q, want %q", c.folder, c.name, got, c.want)
		}
	}
}

func TestParseCaption(t *testing.T) {
	cases := []struct {
		caption        string
		wantFolder     string
		wantName       string
	}{
		{"", "", ""},
		{"photo.png", "", "photo.png"},
		{"#dir:Documents/file.txt", "Documents", "file.txt"},
		{"#dir:Projects/2024/data.csv", "Projects/2024", "data.csv"},
		{"#dir:noslash", "noslash", "noslash"},
	}
	for _, c := range cases {
		folder, name := parseCaption(c.caption)
		if folder != c.wantFolder || name != c.wantName {
			t.Errorf("parseCaption(%q) = (%q,%q), want (%q,%q)",
				c.caption, folder, name, c.wantFolder, c.wantName)
		}
	}
}

func TestTagParseRoundTrip(t *testing.T) {
	pairs := []struct{ folder, name string }{
		{"Documents", "report.pdf"},
		{"Music/Playlist", "song.mp3"},
		{"root", "root.txt"},
	}
	for _, p := range pairs {
		caption := tagCaption(p.folder, p.name)
		folder, name := parseCaption(caption)
		// root maps to empty folder + bare name
		expFolder := p.folder
		if expFolder == "root" {
			expFolder = ""
		}
		if folder != expFolder || name != p.name {
			t.Errorf("round-trip (%q,%q) -> (%q,%q)", p.folder, p.name, folder, name)
		}
	}
}
