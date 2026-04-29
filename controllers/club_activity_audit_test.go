package controllers

import (
	"strings"
	"testing"
)

func TestAuditActivityPublishContentAllowsNormalCopy(t *testing.T) {
	message := auditActivityPublishContent("春季公开试训日", "面向 U10-U14 球员开放，包含基础技术、体能和小场对抗评估。")
	if message != "" {
		t.Fatalf("expected normal activity copy to pass, got %q", message)
	}
}

func TestAuditActivityPublishContentBlocksRiskyCopy(t *testing.T) {
	cases := []struct {
		name        string
		title       string
		description string
		want        string
	}{
		{name: "mobile in title", title: "试训报名 13800000000", want: "活动标题不能包含手机号"},
		{name: "link in description", title: "公开试训", description: "详情见 https://example.com", want: "活动简介不能包含外部链接"},
		{name: "qr code in description", title: "公开试训", description: "到场请扫码添加微信二维码", want: "活动简介不能包含二维码"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			message := auditActivityPublishContent(tc.title, tc.description)
			if message == "" || !strings.Contains(message, tc.want) {
				t.Fatalf("expected message containing %q, got %q", tc.want, message)
			}
		})
	}
}
