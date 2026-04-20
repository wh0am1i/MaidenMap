package geocode

import (
	"strings"

	"github.com/mozillazg/go-pinyin"
)

// pinyinArgs is a package-level arg set (heteronym-aware, no tones, first
// letter uppercase). Shared so we amortize the phrase-dict setup cost.
var pinyinArgs = func() pinyin.Args {
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	a.Heteronym = false
	return a
}()

// chineseAdminSuffixEn translates the canonical suffix of a Chinese admin
// unit to its conventional English form.
//
// Entries that map to "" mean "drop the suffix" — Chinese prefecture cities
// and provinces are commonly written in English without a trailing word (we
// say "Hangzhou", not "Hangzhou City"; "Zhejiang", not "Zhejiang Province").
//
// Suffixes are tried longest-first, so "自治州" matches before "州".
var chineseAdminSuffixEn = []struct {
	zh string
	en string
}{
	{"特别行政区", "Special Administrative Region"},
	{"维吾尔自治区", "Autonomous Region"},
	{"壮族自治区", "Autonomous Region"},
	{"回族自治区", "Autonomous Region"},
	{"自治区", "Autonomous Region"},
	{"自治州", "Autonomous Prefecture"},
	{"自治县", "Autonomous County"},
	{"自治旗", "Autonomous Banner"},
	{"地区", ""},
	{"林区", "Forestry District"},
	{"新区", "New District"},
	{"特区", "Special Zone"},
	{"群岛", "Islands"},
	{"省", ""},
	{"市", ""},
	{"区", "District"},
	{"县", "County"},
	{"旗", "Banner"},
	{"盟", "League"},
	{"州", "Prefecture"},
}

// chineseToEn converts a Chinese admin-area name to a best-effort English
// form: strip the conventional suffix, pinyin-romanize the base, append the
// translated suffix if we use one. Intended for DataV city / district names
// where no curated English is available.
//
// Limits: names whose conventional English is NOT a pinyin transliteration
// (e.g. 乌鲁木齐 → Ürümqi, 呼和浩特 → Hohhot) come out as their pinyin form.
// For admin1 we bypass this entirely via the curated provinceEnByZh table.
func chineseToEn(zh string) string {
	if zh == "" {
		return ""
	}
	for _, s := range chineseAdminSuffixEn {
		if strings.HasSuffix(zh, s.zh) {
			base := strings.TrimSuffix(zh, s.zh)
			en := pinyinJoin(base)
			if en == "" {
				// Degenerate input like "省" alone — fall back to suffix translation only.
				return s.en
			}
			if s.en == "" {
				return en
			}
			return en + " " + s.en
		}
	}
	return pinyinJoin(zh)
}

// heteronymFixups replaces Chinese substrings whose standard place-name
// reading differs from go-pinyin's default single-char reading. Short list —
// only admin-name heteronyms that actually matter. Applied as whole-string
// substring replacement before per-character conversion.
var heteronymFixups = []struct{ zh, en string }{
	{"重庆", "Chongqing"},
	{"六安", "Lu'an"},
	{"乐清", "Yueqing"},
	{"蚌埠", "Bengbu"},
	{"亳州", "Bozhou"},
	{"丽水", "Lishui"},
	{"莆田", "Putian"},
	{"莘县", "Shenxian"},
}

// pinyinJoin converts each character to pinyin and concatenates into a single
// capitalized token — the convention for place names (Hangzhou, Xihu,
// Shijiazhuang). Only the first letter of the whole name is capitalized.
func pinyinJoin(zh string) string {
	if zh == "" {
		return ""
	}

	// Pre-apply heteronym fixups: replace the Chinese phrase with its correct
	// romanization so later pinyin lookup sees only characters it can handle.
	var prefix string
	for _, fx := range heteronymFixups {
		if strings.HasPrefix(zh, fx.zh) {
			prefix = fx.en
			zh = strings.TrimPrefix(zh, fx.zh)
			break
		}
	}

	parts := pinyin.Pinyin(zh, pinyinArgs)
	var b strings.Builder
	b.WriteString(prefix)
	for _, p := range parts {
		if len(p) == 0 || p[0] == "" {
			continue
		}
		b.WriteString(p[0])
	}
	out := b.String()
	if out == "" {
		// No Han characters at all → pass through unchanged (non-Chinese input).
		return zh
	}
	return strings.ToUpper(out[:1]) + out[1:]
}
