// site.go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Site struct {
	Posts  []*Post
	Tags   map[string][]*Post
	Errors []string
}

func LoadSite(opts Options) *Site {
	s := &Site{
		Posts:  make([]*Post, 0),
		Tags:   make(map[string][]*Post),
		Errors: make([]string, 0),
	}

	filepath.Walk(opts.Posts, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.Errors = append(s.Errors, err.Error())
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		post, err := ParsePost(path)
		if err != nil {
		s.Errors = append(s.Errors, err.Error())
		fmt.Printf("FAILED: %s - %v\n", path, err)
		return nil
		}
		fmt.Printf("OK: %s\n", path)

		if post.Draft && !opts.Draft {
			return nil
		}

		s.Posts = append(s.Posts, post)
		for _, tag := range post.Tags {
			s.Tags[tag] = append(s.Tags[tag], post)
		}
		return nil
	})

	// newest first
	sort.Slice(s.Posts, func(i, j int) bool {
		return s.Posts[i].Date.After(s.Posts[j].Date)
	})

	// sort each tag's posts too
	for _, posts := range s.Tags {
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].Date.After(posts[j].Date)
		})
	}

	return s
}
