# ssg - static site generator

A minimal static site generator in Go.

## Build
```
make
```

## Usage

Run from your content directory:
```
cd site-src
../ssg -o ../docs
```

Or from anywhere:
```
ssg -p /path/to/_posts -o /path/to/output
```

### Options
```
-o string    output directory (required)
-p string    posts directory (default "_posts")
-draft       include draft posts
-d           debug output
-v           show version
-h           show help
```

## Content Structure
```
site-src/
  _posts/           # markdown posts
  assets/           # css, js, images
  templates/        # html templates
  content/          # static pages (about.md)
```

### Post Format
```markdown
---
title: My Post
tags: tag1 tag2
draft: true       # optional
---

Summary goes here.
<!--more-->
Full content here.
```

Date comes from filename: `2026-01-25-my-post.md`

## Local Preview
```
cd docs
python3 -m http.server 8000
```

## Dependencies

- goldmark (vendored)
