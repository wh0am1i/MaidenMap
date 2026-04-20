package geocode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChineseToEn(t *testing.T) {
	cases := []struct{ in, want string }{
		// Province / municipality suffixes drop in English.
		{"浙江省", "Zhejiang"},
		{"杭州市", "Hangzhou"},
		{"北京市", "Beijing"},
		{"上海市", "Shanghai"},
		// District / county keep their translated suffix.
		{"西湖区", "Xihu District"},
		{"富阳区", "Fuyang District"},
		{"余杭县", "Yuhang County"},
		// Autonomous units use the long form.
		{"内蒙古自治区", "Neimenggu Autonomous Region"},
		{"广西壮族自治区", "Guangxi Autonomous Region"},
		{"延边朝鲜族自治州", "Yanbianchaoxianzu Autonomous Prefecture"},
		// SAR form (we also have curated English for these in provinceEnByZh,
		// but chineseToEn must still produce a reasonable fallback).
		{"香港特别行政区", "Xianggang Special Administrative Region"},
		// Edge: empty input and pure non-Chinese stay as-is.
		{"", ""},
		{"Taipei", "Taipei"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, chineseToEn(c.in), "input=%q", c.in)
	}
}

func TestPinyinJoinHeteronyms(t *testing.T) {
	// 重 defaults to "zhong" in go-pinyin's single-char dict; our
	// heteronymFixups override produces the correct "Chongqing".
	assert.Equal(t, "Chongqing", pinyinJoin("重庆"))
	assert.Equal(t, "Lu'an", pinyinJoin("六安"))
	// Combined with suffix-stripping it still works end-to-end.
	assert.Equal(t, "Chongqing", chineseToEn("重庆市"))
}
