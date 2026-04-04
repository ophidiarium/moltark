package moltark

import "strings"

const (
	gitattributesBegin = "# BEGIN Moltark managed"
	gitattributesEnd   = "# END Moltark managed"
)

func managedGitattributesBlock() string {
	return strings.Join([]string{
		gitattributesBegin,
		".moltark/** linguist-generated=true",
		gitattributesEnd,
	}, "\n")
}

func mutateGitattributes(raw string) string {
	block := managedGitattributesBlock()
	if strings.TrimSpace(raw) == "" {
		return block + "\n"
	}

	begin, end, ok := managedBlockBounds(raw)
	if ok {
		updated := raw[:begin] + block + raw[end:]
		return ensureTrailingNewline(updated)
	}

	trimmed := strings.TrimRight(raw, "\n")
	return trimmed + "\n\n" + block + "\n"
}

func currentManagedGitattributesBlock(raw string) (string, bool) {
	begin, end, ok := managedBlockBounds(raw)
	if !ok {
		return "", false
	}

	return raw[begin:end], true
}

func managedBlockBounds(raw string) (int, int, bool) {
	begin := strings.Index(raw, gitattributesBegin)
	if begin < 0 {
		return -1, -1, false
	}

	searchStart := begin + len(gitattributesBegin)
	offset := strings.Index(raw[searchStart:], gitattributesEnd)
	if offset < 0 {
		return -1, -1, false
	}

	end := searchStart + offset + len(gitattributesEnd)
	return begin, end, true
}
