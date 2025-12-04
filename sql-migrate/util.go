package migrate

import (
	"regexp"
	"strconv"
	"strings"
)

var regex = regexp.MustCompile("^V.+__.*?\\.sql$")
var digitRegex = regexp.MustCompile("^[1-9][0-9]*$")

func IsSupportFilename(filename string) bool {
	return regex.MatchString(filename)
}

func SplitFilename(filename string) (version string) {
	index := strings.Index(filename, "__")
	return filename[1:index]
}

func CompareVersion(vi, vj string) bool {
	if digitRegex.MatchString(vi) && digitRegex.MatchString(vj) {
		vii, err := strconv.ParseUint(vi, 10, 64)
		if err != nil {
			return compareVersionString(vi, vj)
		}
		vji, err := strconv.ParseUint(vj, 10, 64)
		if err != nil {
			return compareVersionString(vi, vj)
		}
		return vii < vji
	}
	return compareVersionString(vi, vj)
}

func compareVersionString(vi, vj string) bool {
	return vi < vj
}
