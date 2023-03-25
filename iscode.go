package main

import "strings"

func isCode(in string) (value bool) {

	if strings.Contains(in, "func") {
		if strings.Contains(in, "{") {
			value = true
		}
	}

	if strings.Contains(in, "class") {
		if strings.Contains(in, "{") {
			value = true
		} else if strings.Contains(in, "PS>") {
			value = false
		}
	}

	if strings.Contains(in, "println") {
		if strings.Contains(in, "{") {
			value = true
		}
	}

	if strings.Contains(in, "public") {
		if strings.Contains(in, "{") {
			value = true
		}
	}

	if strings.Contains(in, "stdio.h") {
		if strings.Contains(in, "scanf") {
			value = true
		}
	}

	if strings.Contains(in, "{{end}}") {
		value = true
	}

	return value
}
