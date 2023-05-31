package vcsapi

import (
	"strings"

	"github.com/go-git/go-git/v5/utils/diff"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func Diff(srcContent string, dstContent string) []VCSDiffLine {
	var fileLines []VCSDiffLine

	fd := diff.Do(srcContent, dstContent)
	for _, chunk := range fd {
		lines := strings.Split(chunk.Text, "\n")

		switch chunk.Type {
		case diffmatchpatch.DiffEqual:
			for _, line := range lines {
				fileLines = append(fileLines, VCSDiffLine{
					Operation: 0,
					Content:   line,
				})
			}
		case diffmatchpatch.DiffInsert:
			for _, line := range lines {
				fileLines = append(fileLines, VCSDiffLine{
					Operation: 1,
					Content:   line,
				})
			}
		case diffmatchpatch.DiffDelete:
			for _, line := range lines {
				fileLines = append(fileLines, VCSDiffLine{
					Operation: -1,
					Content:   line,
				})
			}
		}
	}

	return fileLines
}
