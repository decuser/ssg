// ssg.go
package main

import "fmt"

func main() {
	opts := parseOptions()

	site := LoadSite(opts)

	for _, err := range site.Errors {
		fmt.Printf("error: %s\n", err)
	}

	fmt.Printf("Loaded %d posts, %d tags\n", len(site.Posts), len(site.Tags))

	if err := Render(site, opts); err != nil {
		fmt.Printf("render error: %v\n", err)
		return
	}

	fmt.Println("done")
}
