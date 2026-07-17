package order

import "testing"

func TestCatalogHasMembershipAndReport(t *testing.T) {
	m, ok := Catalog["prod_membership_monthly"]
	if !ok {
		t.Fatal("缺少会员产品")
	}
	if m.ProductType != "membership" {
		t.Error("会员产品类型应为 membership")
	}
	r, ok := Catalog["prod_deep_report"]
	if !ok {
		t.Fatal("缺少深度报告产品")
	}
	if r.ProductType != "deep_report" {
		t.Error("报告产品类型应为 deep_report")
	}
}

func TestProductNotFound(t *testing.T) {
	_, ok := Catalog["prod_not_exist"]
	if ok {
		t.Error("不存在的产品不应在目录中")
	}
}
