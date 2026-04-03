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

	begin := strings.Index(raw, gitattributesBegin)
	end := strings.Index(raw, gitattributesEnd)
	if begin >= 0 && end >= begin {
		end += len(gitattributesEnd)
		updated := raw[:begin] + block + raw[end:]
		return ensureTrailingNewline(updated)
	}

	trimmed := strings.TrimRight(raw, "\n")
	return trimmed + "\n\n" + block + "\n"
}

func currentManagedGitattributesBlock(raw string) (string, bool) {
	begin := strings.Index(raw, gitattributesBegin)
	end := strings.Index(raw, gitattributesEnd)
	if begin < 0 || end < begin {
		return "", false
	}

	end += len(gitattributesEnd)
	return raw[begin:end], true
}
