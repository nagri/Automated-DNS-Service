package utils

import (
	"strings"
	"unicode/utf8"
)

func reverse(s string) string {
	size := len(s)
	buf := make([]byte, size)
	for start := 0; start < size; {
		r, n := utf8.DecodeRuneInString(s[start:])
		start += n
		utf8.EncodeRune(buf[size-start:], r)
	}
	return string(buf)
}

func GetHostedZoneString(s string) string {
	reverseHostedZone := reverse(s)
	reverseHostedZoneList := strings.Split(reverseHostedZone, ".")
	hostedZone := strings.Join(reverseHostedZoneList[:2], ".")
	hostedZone = reverse(hostedZone) + "."
	return string(hostedZone)
}

func Remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
