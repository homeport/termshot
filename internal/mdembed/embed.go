// Copyright © 2026 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package mdembed

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	startMarker = regexp.MustCompile(`^(\s*)<!--\s*termshot:id=([A-Za-z0-9][A-Za-z0-9._-]*)\s*-->\s*$`)
	endMarker   = regexp.MustCompile(`^\s*<!--\s*/termshot\s*-->\s*$`)
	imageLine   = regexp.MustCompile(`^(\s*)!\[([^\]]*)\]\(([^)]*)\)\s*$`)
	idPattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
)

// ValidateID checks that the embed marker id is present and uses the allowed charset
func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("--id is required when using --embed")
	}

	if !idPattern.MatchString(id) {
		return fmt.Errorf("embed id %q is invalid, use letters, digits, dots, underscores, or hyphens", id)
	}

	return nil
}

// ValidateTarget checks that the markdown file contains exactly one marker for id
func ValidateTarget(markdownPath, id string) error {
	if err := ValidateID(id); err != nil {
		return err
	}

	content, err := os.ReadFile(filepath.Clean(markdownPath))
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	_, err = Apply(string(content), id, "termshot.png", "")
	return err
}

// UpdateFile writes a screenshot image reference into the markdown file at the
// marker <!-- termshot:id=... -->. The image path stored in markdown is relative
// to the markdown file.
func UpdateFile(markdownPath, id, imagePath, alt string) error {
	if err := ValidateID(id); err != nil {
		return err
	}

	absMarkdown, err := filepath.Abs(markdownPath)
	if err != nil {
		return fmt.Errorf("failed to resolve markdown path: %w", err)
	}

	absImage, err := filepath.Abs(imagePath)
	if err != nil {
		return fmt.Errorf("failed to resolve image path: %w", err)
	}

	rel, err := filepath.Rel(filepath.Dir(absMarkdown), absImage)
	if err != nil {
		return fmt.Errorf("failed to compute relative image path: %w", err)
	}

	info, err := os.Stat(absMarkdown)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	content, err := os.ReadFile(filepath.Clean(absMarkdown))
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	updated, err := Apply(string(content), id, filepath.ToSlash(rel), alt)
	if err != nil {
		return err
	}

	// #nosec G703 -- markdown path is the user-provided --embed target
	if err := os.WriteFile(filepath.Clean(absMarkdown), []byte(updated), info.Mode().Perm()); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

// Apply replaces or inserts a markdown image at the embed marker for id
func Apply(content, id, imageRef, alt string) (string, error) {
	if err := ValidateID(id); err != nil {
		return "", err
	}

	lines, nl, trailingNL := splitLines(content)

	start := -1
	indent := ""
	for i, line := range lines {
		match := startMarker.FindStringSubmatch(line)
		if match == nil || match[2] != id {
			continue
		}

		if start != -1 {
			return "", fmt.Errorf("duplicate embed marker <!-- termshot:id=%s --> found", id)
		}

		start = i
		indent = match[1]
	}

	if start == -1 {
		return "", fmt.Errorf("no embed marker <!-- termshot:id=%s --> found", id)
	}

	nextStart := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if startMarker.MatchString(lines[i]) {
			nextStart = i
			break
		}
	}

	end := -1
	for i := start + 1; i < nextStart; i++ {
		if endMarker.MatchString(lines[i]) {
			end = i
			break
		}
	}

	searchUntil := nextStart
	if end != -1 {
		searchUntil = end
	}

	imageIdx := -1
	existingAlt := ""
	for i := start + 1; i < searchUntil; i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}

		if match := imageLine.FindStringSubmatch(lines[i]); match != nil {
			imageIdx = i
			existingAlt = match[2]
			if indent == "" {
				indent = match[1]
			}
		}

		break
	}

	if end != -1 && existingAlt == "" {
		for i := start + 1; i < end; i++ {
			if match := imageLine.FindStringSubmatch(lines[i]); match != nil {
				existingAlt = match[2]
				break
			}
		}
	}

	if alt == "" {
		alt = existingAlt
	}

	if alt == "" {
		alt = id
	}

	image := indent + "![" + alt + "](" + imageRef + ")"

	var out []string
	switch {
	case end != -1:
		out = append(out, lines[:start+1]...)
		out = append(out, image)
		out = append(out, lines[end:]...)

	case imageIdx != -1:
		lines[imageIdx] = image
		out = lines

	default:
		out = append(out, lines[:start+1]...)
		out = append(out, image)
		out = append(out, lines[start+1:]...)
	}

	result := strings.Join(out, nl)
	if trailingNL {
		result += nl
	}

	return result, nil
}

func splitLines(content string) ([]string, string, bool) {
	nl := "\n"
	if strings.Contains(content, "\r\n") {
		nl = "\r\n"
		content = strings.ReplaceAll(content, "\r\n", "\n")
	}

	trailingNL := strings.HasSuffix(content, "\n")
	content = strings.TrimSuffix(content, "\n")
	if content == "" {
		return []string{}, nl, trailingNL
	}

	return strings.Split(content, "\n"), nl, trailingNL
}
