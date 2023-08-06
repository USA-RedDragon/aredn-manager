package utils

import "strings"

func ShellReplace(src *string, variables map[string]string) {
	srcCpy := *src
	for key, value := range variables {
		srcCpy = strings.ReplaceAll(srcCpy, "${"+key+"}", value)
	}
	*src = srcCpy
}
