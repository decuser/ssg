// post.go
package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

type Post struct {
	Title   string
	Date    time.Time
	Tags    []string
	Slug    string
	Summary string
	Content string
	Path    string
	Draft   bool
}

func ParsePost(path string) (*Post, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	front, body := splitFrontMatter(data)
	title, date, tags, draft := parseFrontMatter(front)
	summary, content := splitContent(body)
	slug := slugFromPath(path)

	// If date is zero, try to get it from filename
	if date.IsZero() {
		date = dateFromPath(path)
	}

	return &Post{
		Title:   title,
		Date:    date,
		Tags:    tags,
		Slug:    slug,
		Summary: summary,
		Content: content,
		Path:    path,
		Draft:   draft,
	}, nil
}

func dateFromPath(path string) time.Time {
	base := filepath.Base(path)
	// filename format: 2023-01-16-title.md
	if len(base) >= 10 {
		dateStr := base[:10]
		date, _ := time.Parse("2006-01-02", dateStr)
		return date
	}
	return time.Time{}
}

func splitFrontMatter(data []byte) (front []byte, body []byte) {
	s := bufio.NewScanner(bytes.NewReader(data))
	inFront := false
	var frontLines, bodyLines []string

	for s.Scan() {
		line := s.Text()
		if line == "---" {
			if !inFront {
				inFront = true
				continue
			} else {
				inFront = false
				continue
			}
		}
		if inFront {
			frontLines = append(frontLines, line)
		} else {
			bodyLines = append(bodyLines, line)
		}
	}
	return []byte(strings.Join(frontLines, "\n")), []byte(strings.Join(bodyLines, "\n"))
}

func parseFrontMatter(data []byte) (title string, date time.Time, tags []string, draft bool) {
	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "title:") {
			title = strings.TrimSpace(strings.TrimPrefix(line, "title:"))
			title = strings.Trim(title, "\"'")
		} else if strings.HasPrefix(line, "date:") {
			dateStr := strings.TrimSpace(strings.TrimPrefix(line, "date:"))
			date, _ = time.Parse("2006-01-02 15:04:05 -0700", dateStr)
			if date.IsZero() {
				date, _ = time.Parse("2006-01-02", dateStr)
			}
		} else if strings.HasPrefix(line, "tags:") {
			tagStr := strings.TrimSpace(strings.TrimPrefix(line, "tags:"))
			tagStr = strings.Trim(tagStr, "[]")
			// comma-separated or space-separated
			if strings.Contains(tagStr, ",") {
				for _, t := range strings.Split(tagStr, ",") {
					t = strings.TrimSpace(t)
					if t != "" {
						tags = append(tags, t)
					}
				}
			} else {
				for _, t := range strings.Fields(tagStr) {
					if t != "" {
						tags = append(tags, strings.ToLower(t))
					}
				}
			}
		} else if strings.HasPrefix(line, "draft:") {
			draftStr := strings.TrimSpace(strings.TrimPrefix(line, "draft:"))
			draft = draftStr == "true"
		}
	}
	return
}

func splitContent(body []byte) (summary string, content string) {
	bodyStr := string(body)

	// Strip Jekyll liquid tags
	bodyStr = strings.ReplaceAll(bodyStr, "{% raw %}", "")
	bodyStr = strings.ReplaceAll(bodyStr, "{% endraw %}", "")

	var summaryMd, contentMd string
	if parts := strings.SplitN(bodyStr, "<!--more-->", 2); len(parts) == 2 {
		summaryMd = parts[0]
		contentMd = bodyStr
	} else {
		summaryMd = bodyStr
		contentMd = bodyStr
	}

	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	md.Convert([]byte(summaryMd), &buf)
	summary = buf.String()

	buf.Reset()
	md.Convert([]byte(strings.ReplaceAll(contentMd, "<!--more-->", "")), &buf)
	content = buf.String()

	return
}

func slugFromPath(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if len(base) > 11 && base[4] == '-' && base[7] == '-' && base[10] == '-' {
		base = base[11:]
	}
	return base
}
