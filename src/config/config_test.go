package config

import "testing"

func TestNewSetConfig_DeepCopy(t *testing.T) {
	set1 := NewSetConfig()
	set2 := NewSetConfig()

	set1.TCP.WinValues = append(set1.TCP.WinValues, 999)
	set1.Targets.GeoSiteCategories = append(set1.Targets.GeoSiteCategories, "test")
	set1.Faking.SNIMutation.FakeSNIs = append(set1.Faking.SNIMutation.FakeSNIs, "test.com")

	if len(set2.TCP.WinValues) != 4 {
		t.Error("WinValues leaked between instances")
	}
	if len(set2.Targets.GeoSiteCategories) != 0 {
		t.Error("GeoSiteCategories leaked between instances")
	}
	if len(set2.Faking.SNIMutation.FakeSNIs) != 3 {
		t.Error("FakeSNIs leaked between instances")
	}
}

func TestNewConfig_DeepCopy(t *testing.T) {
	cfg1 := NewConfig()
	cfg2 := NewConfig()

	cfg1.MainSet.TCP.ConnBytesLimit = 999
	cfg1.System.Checker.Domains = append(cfg1.System.Checker.Domains, "test.com")

	if cfg2.MainSet.TCP.ConnBytesLimit == 999 {
		t.Error("MainSet leaked between instances")
	}
	if len(cfg2.System.Checker.Domains) != 0 {
		t.Error("Checker.Domains leaked between instances")
	}
}
