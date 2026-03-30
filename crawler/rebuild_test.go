package crawler

import (
	"testing"
)

func TestClassifyAction(t *testing.T) {
	tests := []struct {
		name     string
		rptType  string
		expected string
	}{
		{name: "신설 키워드", rptType: "정비구역 신설", expected: "신설"},
		{name: "지정 키워드", rptType: "정비구역 지정", expected: "신설"},
		{name: "설립 키워드", rptType: "조합 설립인가", expected: "신설"},
		{name: "변경 키워드", rptType: "정비계획 변경", expected: "변경"},
		{name: "수정 키워드", rptType: "계획 수정", expected: "변경"},
		{name: "폐지 키워드", rptType: "정비구역 폐지", expected: "폐지"},
		{name: "해제 키워드", rptType: "구역 해제", expected: "폐지"},
		{name: "기타", rptType: "기타 유형", expected: "기타"},
		{name: "빈 문자열", rptType: "", expected: "기타"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyAction(tt.rptType)
			if got != tt.expected {
				t.Errorf("classifyAction(%q) = %q, want %q", tt.rptType, got, tt.expected)
			}
		})
	}
}

func TestStrPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantNil  bool
	}{
		{name: "빈 문자열은 nil", input: "", wantNil: true},
		{name: "값 있으면 포인터", input: "강남구", wantNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strPtr(tt.input)
			if tt.wantNil && got != nil {
				t.Errorf("strPtr(%q) = %v, want nil", tt.input, got)
			}
			if !tt.wantNil {
				if got == nil {
					t.Errorf("strPtr(%q) = nil, want non-nil", tt.input)
				} else if *got != tt.input {
					t.Errorf("strPtr(%q) = %q, want %q", tt.input, *got, tt.input)
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substrs  []string
		expected bool
	}{
		{name: "포함됨", s: "정비구역 신설", substrs: []string{"신설"}, expected: true},
		{name: "포함 안 됨", s: "정비구역 신설", substrs: []string{"변경"}, expected: false},
		{name: "여러 키워드 중 하나 매칭", s: "구역 해제", substrs: []string{"폐지", "해제"}, expected: true},
		{name: "빈 문자열", s: "", substrs: []string{"신설"}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substrs...)
			if got != tt.expected {
				t.Errorf("contains(%q, %v) = %v, want %v", tt.s, tt.substrs, got, tt.expected)
			}
		})
	}
}
