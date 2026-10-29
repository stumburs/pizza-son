package webui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestPagesRender(t *testing.T) {
	pages := []struct {
		name string
		page templ.Component
	}{
		{"overview", OverviewPage()},
		{"berts", BertsPage()},
		{"bertle", BertlePage()},
		{"quotes", QuotesPage()},
		{"docs", DocsPage()},
		{"vote", VotePage()},
		{"panel", PanelPage()},
	}

	for _, tc := range pages {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tc.page.Render(context.Background(), &buf); err != nil {
				t.Fatalf("render: %v", err)
			}

			html := buf.String()
			if !strings.HasPrefix(html, "<!doctype html>") {
				t.Fatalf("unexpected document start: %.40q", html)
			}
			for _, want := range []string{
				`class="topbar"`,
				"/css/site.css",
				"/js/" + tc.name + ".js",
				`class="site-main page-` + tc.name + `"`,
			} {
				if !strings.Contains(html, want) {
					t.Errorf("missing %q", want)
				}
			}
			if strings.Contains(html, "/views/") || strings.Contains(html, "/view.js") {
				t.Errorf("legacy iframe assets are still referenced")
			}
		})
	}
}
