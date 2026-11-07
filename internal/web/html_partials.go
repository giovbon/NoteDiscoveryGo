package web

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

// htmlIncludePattern matches a whole-line include directive:
//
//	<!-- @include partials/sidebar/files.html -->
//
// The directive must occupy its own line. The optional newline is part of the
// match so a partial that ends with a newline does not leave an extra blank
// line behind.
var htmlIncludePattern = regexp.MustCompile(`(?m)^[ \t]*<!--[ \t]*@include[ \t]+([^ \t>]+)[ \t]*-->[ \t]*(\r?\n)?`)

// maxHTMLIncludeDepth guards against pathological (or cyclic) include trees.
const maxHTMLIncludeDepth = 32

// expandHTMLIncludes replaces <!-- @include path --> directives in content
// with the file at path, resolved relative to frontend/. Includes are read from
// assetsFS so the embedded build and the on-disk web/ override behave the same.
//
// Partials are expanded once at startup by buildIndex, not per request.
func expandHTMLIncludes(assetsFS fs.FS, content string) (string, error) {
	return expandHTMLIncludesTree(assetsFS, content, nil, 0)
}

func expandHTMLIncludesTree(assetsFS fs.FS, content string, stack []string, depth int) (string, error) {
	if depth > maxHTMLIncludeDepth {
		return "", fmt.Errorf("html includes nested deeper than %d levels", maxHTMLIncludeDepth)
	}

	var firstErr error
	expanded := htmlIncludePattern.ReplaceAllStringFunc(content, func(directive string) string {
		if firstErr != nil {
			return directive
		}

		match := htmlIncludePattern.FindStringSubmatch(directive)
		if match == nil {
			return directive
		}
		includePath := match[1]
		hasNewline := match[2] != ""

		clean := path.Clean(includePath)
		if clean == "." || clean == ".." || path.IsAbs(clean) || strings.HasPrefix(clean, "../") {
			firstErr = fmt.Errorf("html include escapes frontend/: %q", includePath)
			return directive
		}

		fullPath := path.Join("frontend", clean)
		for _, seen := range stack {
			if seen == fullPath {
				chain := append(append([]string{}, stack...), fullPath)
				firstErr = fmt.Errorf("html include cycle: %s", strings.Join(chain, " -> "))
				return directive
			}
		}

		data, err := fs.ReadFile(assetsFS, fullPath)
		if err != nil {
			firstErr = fmt.Errorf("html include %q: %w", includePath, err)
			return directive
		}

		nextStack := append(append([]string{}, stack...), fullPath)
		out, err := expandHTMLIncludesTree(assetsFS, string(data), nextStack, depth+1)
		if err != nil {
			firstErr = err
			return directive
		}

		// A partial is a block: if the directive ended the line, keep the
		// replacement ending on a line boundary too.
		if hasNewline && !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		return out
	})
	if firstErr != nil {
		return "", firstErr
	}
	return expanded, nil
}
