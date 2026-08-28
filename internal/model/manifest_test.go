package model

import "testing"

func TestManifestPackagesForPlatformFiltersPackages(t *testing.T) {
	manifest := Manifest{Version: 1, Packages: []Package{
		{ID: "jdk", Name: "JDK", Manager: "winget", ManagerID: "jdk", Platforms: []Platform{PlatformWindows, PlatformDarwin}},
		{ID: "brew-only", Name: "brew", Manager: "brew", ManagerID: "brew-only", Platforms: []Platform{PlatformDarwin}},
	}}

	got, err := manifest.PackagesForPlatform(PlatformWindows)
	if err != nil {
		t.Fatalf("筛选清单失败: %v", err)
	}
	if len(got) != 1 || got[0].ID != "jdk" {
		t.Fatalf("筛选结果错误: %#v", got)
	}
}

func TestManifestValidateRejectsEmptyPackageID(t *testing.T) {
	err := (Manifest{Version: 1, Packages: []Package{{Name: "没有 ID"}}}).Validate()
	if err == nil {
		t.Fatal("空包 ID 应该校验失败")
	}
}
