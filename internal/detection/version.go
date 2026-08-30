package detection

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var versionPattern = regexp.MustCompile(`\d+(?:\.\d+)*`)

func MajorVersion(output string) (int, error) {
	version, err := VersionText(output)
	if err != nil {
		return 0, err
	}
	majorText := version
	if index := strings.IndexByte(version, '.'); index >= 0 {
		majorText = version[:index]
	}
	major, err := strconv.Atoi(majorText)
	if err != nil {
		return 0, fmt.Errorf("解析主版本失败: %w", err)
	}
	return major, nil
}

func VersionText(output string) (string, error) {
	version := versionPattern.FindString(output)
	if version == "" {
		return "", fmt.Errorf("版本输出中没有数字: %q", output)
	}
	return version, nil
}

func MeetsMinimum(actual, minimum int) bool {
	return minimum <= 0 || actual >= minimum
}
