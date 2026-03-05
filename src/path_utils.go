package plugin

import (
	"encoding/hex"
	"sort"
	"strings"
)

func normalizeInline(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r", ""))
}

func NormalizePath(base string, raw string) string {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, "\\", "/"))
	if raw == "" {
		if base == "" {
			return "/"
		}
		raw = base
	}
	if !strings.HasPrefix(raw, "/") {
		if base == "" {
			base = "/"
		}
		raw = JoinPath(base, raw)
	}
	parts := strings.Split(raw, "/")
	stack := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			continue
		}
		stack = append(stack, part)
	}
	if len(stack) == 0 {
		return "/"
	}
	return "/" + strings.Join(stack, "/")
}

func JoinPath(base string, name string) string {
	base = NormalizePath("/", base)
	name = strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	if strings.HasPrefix(name, "/") {
		return NormalizePath("/", name)
	}
	if base == "/" {
		return NormalizePath("/", "/"+name)
	}
	return NormalizePath("/", base+"/"+name)
}

func ParentDir(path string) string {
	path = NormalizePath("/", path)
	if path == "/" {
		return "/"
	}
	idx := strings.LastIndex(path, "/")
	if idx <= 0 {
		return "/"
	}
	return path[:idx]
}

func BaseName(path string) string {
	path = NormalizePath("/", path)
	if path == "/" {
		return "root"
	}
	idx := strings.LastIndex(path, "/")
	if idx < 0 || idx == len(path)-1 {
		return path
	}
	return path[idx+1:]
}

func ShellQuote(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "''"
	}
	safe := true
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			continue
		}
		switch ch {
		case '/', '.', '-', '_':
			continue
		default:
			safe = false
		}
		if !safe {
			break
		}
	}
	if safe {
		return s
	}
	// nsh 对双引号转义兼容性差，统一改为单引号包裹。
	// 对极少出现的单引号字符采用 POSIX 方式转义。
	s = strings.ReplaceAll(s, "'", "'\\''")
	return "'" + s + "'"
}

func ParseLsOutput(output string) []string {
	lines := strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n")
	seen := make(map[string]bool)
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" || name == "." || name == ".." {
			continue
		}
		// Ignore 'total x...' outputs from ls
		if strings.HasPrefix(name, "total ") {
			continue
		}
		// Ignore directory headers like '/data/quickapp/:'
		if strings.HasSuffix(name, ":") {
			continue
		}
		if !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

func BytesToHexPreview(data []byte, maxBytes int) string {
	if maxBytes <= 0 {
		maxBytes = HexPreviewBytes
	}
	if len(data) > maxBytes {
		data = data[:maxBytes]
	}
	if len(data) == 0 {
		return ""
	}
	encoded := strings.ToUpper(hex.EncodeToString(data))
	var builder strings.Builder
	for i := 0; i < len(encoded); i += 2 {
		if i > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(encoded[i : i+2])
	}
	return builder.String()
}

func IsUnderDataRoot(path string) bool {
	path = NormalizePath("/", path)
	return path == "/data" || strings.HasPrefix(path, "/data/")
}
