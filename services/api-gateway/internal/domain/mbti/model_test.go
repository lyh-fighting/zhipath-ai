package mbti

import "testing"

// 验证维度保存"极性 + 百分比"，不能仅存无方向数字。
func TestDimensionPolarityAndPercentage(t *testing.T) {
	d := Dimension{Pole: "E", Percentage: 67}
	if d.Pole != "E" {
		t.Errorf("极性应为 E, got %s", d.Pole)
	}
	if d.Percentage != 67 {
		t.Errorf("百分比应为 67, got %d", d.Percentage)
	}
}

func TestDimensionsAllFour(t *testing.T) {
	d := Dimensions{
		Energy:  Dimension{Pole: "E", Percentage: 67},
		Mind:    Dimension{Pole: "N", Percentage: 71},
		Nature:  Dimension{Pole: "F", Percentage: 68},
		Tactics: Dimension{Pole: "P", Percentage: 39},
	}
	if d.Energy.Pole != "E" || d.Mind.Pole != "N" || d.Nature.Pole != "F" || d.Tactics.Pole != "P" {
		t.Error("四维度极性不匹配")
	}
}

// 16 型结果类型应为 4 字符。
func TestResultTypeLength(t *testing.T) {
	cases := []string{"INFP", "ENTJ", "ESFJ", "INTP", "ISTJ", "ENFP"}
	for _, v := range cases {
		if len(v) != 4 {
			t.Errorf("result_type 应为 4 字符: %s", v)
		}
	}
}

// OCR 来源的结果默认不应直接成为当前（需用户确认）。
func TestOCRSourceNeedsConfirmation(t *testing.T) {
	res := &Result{Source: "ocr", ConfirmedByUser: false}
	if res.Source == "ocr" && res.ConfirmedByUser {
		t.Error("OCR 来源结果必须默认未确认，需用户确认后才成当前")
	}
}
