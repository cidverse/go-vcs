package vcs

import (
	"strings"

	"github.com/gosimple/slug"
)

func getReleaseName(input string) string {
	input = slug.Substitute(input, map[string]string{
		"/": "-",
	})

	return strings.TrimLeft(input, "v")
}
