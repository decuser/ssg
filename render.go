// render.go
package main

import (
	"bytes"
	"encoding/json"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type TagInfo struct {
	Name  string
	Count int
}

type PageData struct {
	Title      string
	Content    template.HTML
	Tags       []TagInfo
	Posts      []*Post
	Post       *Post
	TagName    string
	Page       int
	TotalPages int
	PrevPage   string
	NextPage   string
	Years      []YearGroup
}

type YearGroup struct {
	Year   int
	Months []MonthGroup
}

type MonthGroup struct {
	Month string
	Posts []*Post
}

func loadTemplate(content string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	return template.New("").Funcs(funcMap).ParseFiles("templates/base.html", "templates/"+content)
}

func Render(site *Site, opts Options) error {
	os.RemoveAll(opts.Output)

	tagInfo := buildTagInfo(site)

	if err := renderIndex(site, tagInfo, opts); err != nil {
		return err
	}

	if err := renderPosts(site, tagInfo, opts); err != nil {
		return err
	}

	if err := renderTags(site, tagInfo, opts); err != nil {
		return err
	}

	if err := renderAll(site, tagInfo, opts); err != nil {
		return err
	}

	if err := renderAbout(tagInfo, opts); err != nil {
		return err
	}

	if err := copyAssets(opts); err != nil {
		return err
	}

	if err := renderSearchJSON(site, opts); err != nil {
		return err
	}

	if err := renderRSS(site, opts); err != nil {
		return err
	}

	// create .nojekyll
	os.WriteFile(filepath.Join(opts.Output, ".nojekyll"), []byte{}, 0644)

	return nil
}

func buildTagInfo(site *Site) []TagInfo {
	var tags []TagInfo
	for name, posts := range site.Tags {
		tags = append(tags, TagInfo{Name: name, Count: len(posts)})
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name < tags[j].Name
	})
	return tags
}

func renderIndex(site *Site, tagInfo []TagInfo, opts Options) error {
	tmpl, err := loadTemplate("index.html")
	if err != nil {
		return err
	}

	perPage := 10
	total := len(site.Posts)
	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	for page := 1; page <= totalPages; page++ {
		start := (page - 1) * perPage
		end := start + perPage
		if end > total {
			end = total
		}

		var posts []*Post
		if total > 0 {
			posts = site.Posts[start:end]
		}

		data := PageData{
			Title:      "home",
			Tags:       tagInfo,
			Posts:      posts,
			Page:       page,
			TotalPages: totalPages,
		}

		if page > 1 {
			if page == 2 {
				data.PrevPage = "/"
			} else {
				data.PrevPage = "/page/" + strconv.Itoa(page-1) + "/"
			}
		}
		if page < totalPages {
			data.NextPage = "/page/" + strconv.Itoa(page+1) + "/"
		}

		var outPath string
		if page == 1 {
			outPath = filepath.Join(opts.Output, "index.html")
		} else {
			outPath = filepath.Join(opts.Output, "page", strconv.Itoa(page), "index.html")
		}

		if err := writeTemplate(tmpl, outPath, data); err != nil {
			return err
		}
	}
	return nil
}

func renderPosts(site *Site, tagInfo []TagInfo, opts Options) error {
	tmpl, err := loadTemplate("post.html")
	if err != nil {
		return err
	}

	for _, post := range site.Posts {
		data := PageData{
			Title: post.Title,
			Tags:  tagInfo,
			Post:  post,
		}
		outPath := filepath.Join(opts.Output, "posts", post.Slug, "index.html")
		if err := writeTemplate(tmpl, outPath, data); err != nil {
			return err
		}
	}
	return nil
}

func renderTags(site *Site, tagInfo []TagInfo, opts Options) error {
	tmpl, err := loadTemplate("tag.html")
	if err != nil {
		return err
	}

	perPage := 10

	for tag, posts := range site.Tags {
		total := len(posts)
		totalPages := (total + perPage - 1) / perPage

		for page := 1; page <= totalPages; page++ {
			start := (page - 1) * perPage
			end := start + perPage
			if end > total {
				end = total
			}

			data := PageData{
				Title:      tag,
				Tags:       tagInfo,
				Posts:      posts[start:end],
				TagName:    tag,
				Page:       page,
				TotalPages: totalPages,
			}

			if page > 1 {
				if page == 2 {
					data.PrevPage = "/tags/" + tag + "/"
				} else {
					data.PrevPage = "/tags/" + tag + "/page/" + strconv.Itoa(page-1) + "/"
				}
			}
			if page < totalPages {
				data.NextPage = "/tags/" + tag + "/page/" + strconv.Itoa(page+1) + "/"
			}

			var outPath string
			if page == 1 {
				outPath = filepath.Join(opts.Output, "tags", tag, "index.html")
			} else {
				outPath = filepath.Join(opts.Output, "tags", tag, "page", strconv.Itoa(page), "index.html")
			}

			if err := writeTemplate(tmpl, outPath, data); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderAll(site *Site, tagInfo []TagInfo, opts Options) error {
	tmpl, err := loadTemplate("all.html")
	if err != nil {
		return err
	}

	years := buildYearGroups(site.Posts)
	data := PageData{
		Title: "all posts",
		Tags:  tagInfo,
		Years: years,
	}
	outPath := filepath.Join(opts.Output, "all", "index.html")
	return writeTemplate(tmpl, outPath, data)
}

func buildYearGroups(posts []*Post) []YearGroup {
	yearMap := make(map[int]map[string][]*Post)

	for _, p := range posts {
		y := p.Date.Year()
		m := p.Date.Format("January")
		if yearMap[y] == nil {
			yearMap[y] = make(map[string][]*Post)
		}
		yearMap[y][m] = append(yearMap[y][m], p)
	}

	var years []int
	for y := range yearMap {
		years = append(years, y)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))

	var result []YearGroup
	for _, y := range years {
		monthMap := yearMap[y]
		var months []MonthGroup
		monthOrder := []string{"January", "February", "March", "April", "May", "June",
			"July", "August", "September", "October", "November", "December"}
		for i := len(monthOrder) - 1; i >= 0; i-- {
			m := monthOrder[i]
			if posts, ok := monthMap[m]; ok {
				months = append(months, MonthGroup{Month: m, Posts: posts})
			}
		}
		result = append(result, YearGroup{Year: y, Months: months})
	}
	return result
}

func renderAbout(tagInfo []TagInfo, opts Options) error {
	tmpl, err := loadTemplate("about.html")
	if err != nil {
		return err
	}

	data, err := os.ReadFile("content/about.md")
	if err != nil {
		return err
	}

	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	md.Convert(data, &buf)

	pageData := PageData{
		Title:   "about",
		Tags:    tagInfo,
		Content: template.HTML(buf.String()),
	}
	outPath := filepath.Join(opts.Output, "about", "index.html")
	return writeTemplate(tmpl, outPath, pageData)
}

func writeTemplate(tmpl *template.Template, outPath string, data PageData) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.ExecuteTemplate(f, "base.html", data)
}

func copyAssets(opts Options) error {
	src := "assets"
	dst := filepath.Join(opts.Output, "assets")
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

func renderSearchJSON(site *Site, opts Options) error {
	type SearchEntry struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Date  string `json:"date"`
		Tags  string `json:"tags"`
	}

	var entries []SearchEntry
	for _, post := range site.Posts {
		entries = append(entries, SearchEntry{
			Title: post.Title,
			URL:   "/posts/" + post.Slug + "/",
			Date:  post.Date.Format("Jan 2, 2006"),
			Tags:  strings.Join(post.Tags, " "),
		})
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	outPath := filepath.Join(opts.Output, "search.json")
	return os.WriteFile(outPath, data, 0644)
}
