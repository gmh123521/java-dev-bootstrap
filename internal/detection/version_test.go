package detection

import "testing"

func TestMajorVersionParsesJavaAndToolOutputs(t *testing.T) {
	cases := map[string]int{
		`java version "23.0.2"`:            23,
		`openjdk version "21.0.5"`:         21,
		`git version 2.51.1.windows.1`:     2,
		`Apache Maven 3.9.9`:               3,
		`Gradle 8.12`:                      8,
		`Docker version 27.5.1, build abc`: 27,
	}

	for output, expected := range cases {
		actual, err := MajorVersion(output)
		if err != nil {
			t.Fatalf("解析 %q 失败: %v", output, err)
		}
		if actual != expected {
			t.Fatalf("解析 %q 得到 %d，期望 %d", output, actual, expected)
		}
	}
}

func TestMeetsMinimumAcceptsNewerJava(t *testing.T) {
	if !MeetsMinimum(23, 21) {
		t.Fatal("Java 23 应满足最低 Java 21")
	}
	if MeetsMinimum(17, 21) {
		t.Fatal("Java 17 不应满足最低 Java 21")
	}
	if !MeetsMinimum(1, 0) {
		t.Fatal("未设置最低版本时应视为满足")
	}
}

func TestVersionTextReturnsFullVersion(t *testing.T) {
	actual, err := VersionText(`openjdk version "21.0.5" 2024-10-15`)
	if err != nil {
		t.Fatalf("提取完整版本失败: %v", err)
	}
	if actual != "21.0.5" {
		t.Fatalf("完整版本为 %q，期望 21.0.5", actual)
	}
}
