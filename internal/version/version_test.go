package version

import "testing"

func TestDisplayReturnsDevelopmentVersionByDefault(t *testing.T) {
	if Display() == "" {
		t.Fatal("版本显示不能为空")
	}
}
