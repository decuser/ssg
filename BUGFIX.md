# Bugfix: Duplicate Tags and Duplicate Posts on Tag Pages

**Date:** 2026-08-15

## Symptom

A post appears multiple times on a tag's page, and the same tag appears
more than once on a post. Visible in prod with the alpha-milestone-26
update post.

## Root Cause

`splitFrontMatter` in `post.go` toggles `inFront` on every `---` line in
the file. The milestone-26 post contains a fenced markdown code block that
quotes its own frontmatter as an example:

    ````markdown
    ---
    title: Aiki: Alpha Milestone 26 Update
    tags: aiki programming-languages ai self-hosting
    ---

The `---` lines inside the code block re-enter frontmatter mode, so
`parseFrontMatter` sees the `tags:` line twice. Each duplicate tag causes
`site.go` to append the post to that tag's list again.

## Fix

Two changes in `post.go`:

1. **`splitFrontMatter`**: Add a `frontDone` flag. Once the closing `---`
   is found, all subsequent `---` lines are treated as body content.

2. **`parseFrontMatter`** (minor): The comma-separated tag branch was not
   lowercasing tags, unlike the space-separated branch. Normalized both
   branches to `ToLower` and added a `seen` map to deduplicate within a
   single post's tag list.
