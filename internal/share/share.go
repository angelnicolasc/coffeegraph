// Package share publishes job logs as GitHub gists or local HTML receipts.
package share

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/coffeegraph/coffeegraph/internal/fsutil"
	"github.com/coffeegraph/coffeegraph/internal/logs"
)

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)sk-ant-[A-Za-z0-9_-]+`),
	regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9._-]+`),
	regexp.MustCompile(`(?i)TELEGRAM_[A-Z0-9_]+=?.*`),
	regexp.MustCompile(`(?i)GITHUB_TOKEN=?.*`),
}

// Result contains a share destination.
type Result struct {
	URL      string
	Local    string
	Fallback bool
}

// Publish creates a share link for a log entry.
func Publish(entry logs.Entry, githubToken string, public bool, projectRoot string) (Result, error) {
	md := formatMarkdown(entry)
	if strings.TrimSpace(githubToken) != "" {
		u, err := createGist(githubToken, entry, md, public)
		if err == nil {
			return Result{URL: u}, nil
		}
	}
	p, err := writeHTML(projectRoot, entry, md)
	if err != nil {
		return Result{}, err
	}
	_ = openInBrowser(p)
	return Result{Local: p, Fallback: true}, nil
}

// Sanitize removes lines containing common secrets.
func Sanitize(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		redact := false
		for _, p := range secretPatterns {
			if p.MatchString(line) {
				redact = true
				break
			}
		}
		if !redact {
			out = append(out, line)
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func formatMarkdown(e logs.Entry) string {
	res := Sanitize(e.Result)
	if strings.TrimSpace(e.Task) == "" {
		e.Task = "(not recorded)"
	}
	return fmt.Sprintf(`# CoffeeGraph Result

**Skill:** %s  
**Task:** %s  
**Timestamp:** %s

## Output

%s
`, e.Skill, e.Task, e.FinishedAt.Local().Format(time.RFC1123), res)
}

func createGist(token string, e logs.Entry, md string, public bool) (string, error) {
	body := map[string]any{
		"description": fmt.Sprintf("CoffeeGraph job %s (%s)", e.ID, e.Skill),
		"public":      public,
		"files": map[string]map[string]string{
			fmt.Sprintf("coffeegraph-%s.md", e.ID): {"content": md},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.github.com/gists", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("github gist api %s", resp.Status)
	}
	var parsed struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	return parsed.HTMLURL, nil
}

func writeHTML(root string, e logs.Entry, md string) (string, error) {
	ts := time.Now().Format("20060102-150405")
	outDir := filepath.Join(root, ".coffee", "share")
	p := filepath.Join(outDir, ts+".html")
	card := fmt.Sprintf(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>CoffeeGraph Share</title>
<style>
body{font-family:ui-sans-serif,system-ui,-apple-system,Segoe UI,sans-serif;background:linear-gradient(140deg,#f8f5ef,#f2ede4);margin:0;padding:32px;color:#272019}
.card{max-width:860px;margin:0 auto;background:#fff;border:1px solid #e7dfd3;border-radius:16px;box-shadow:0 14px 40px rgba(74,58,37,.1);padding:28px}
h1{margin:0 0 6px;font-size:32px;color:#6b4e2f}.meta{color:#6f6556;margin-bottom:24px}
pre{white-space:pre-wrap;background:#f9f7f2;border:1px solid #ece4d8;border-radius:12px;padding:16px}
</style></head><body><div class="card"><h1>%s</h1><div class="meta">%s</div><pre>%s</pre><div class="meta">%s</div></div></body></html>`,
		htmlEsc(e.Skill), htmlEsc(e.Task), htmlEsc(Sanitize(e.Result)), htmlEsc(e.FinishedAt.Local().Format(time.RFC1123)))
	if err := fsutil.AtomicWriteFile(p, []byte(card), 0644); err != nil {
		return "", err
	}
	return p, nil
}

func openInBrowser(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	switch {
	case isWindows():
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}

func isWindows() bool { return strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") }

func htmlEsc(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return r.Replace(s)
}
