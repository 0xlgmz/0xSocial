package httpapi

import "testing"

func TestValidReportReason(t *testing.T) {
	valid := []string{
		"spam",
		"harassment",
		"hate_speech",
		"violence",
		"sexual_content",
		"self_harm",
		"false_information",
		"other",
	}
	for _, reason := range valid {
		if !validReportReason(reason) {
			t.Errorf("validReportReason(%q) = false, want true", reason)
		}
	}

	invalid := []string{"", "SPAM", "misinformation", "other_reason"}
	for _, reason := range invalid {
		if validReportReason(reason) {
			t.Errorf("validReportReason(%q) = true, want false", reason)
		}
	}
}
