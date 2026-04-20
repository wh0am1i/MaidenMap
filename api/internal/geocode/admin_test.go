package geocode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveAdminNames(t *testing.T) {
	a1 := map[string]AdminEntry{
		"US.CA": {En: "California", Zh: "加利福尼亚"},
		"DE.BR": {En: "Brandenburg", Zh: "勃兰登堡"},
	}
	a2 := map[string]AdminEntry{
		"US.CA.037": {En: "Los Angeles", Zh: "洛杉矶"},
	}

	n1, n2 := ResolveAdminNames(a1, a2, "US", "CA", "037")
	assert.Equal(t, "California", n1.En)
	assert.Equal(t, "加利福尼亚", n1.Zh)
	assert.Equal(t, "Los Angeles", n2.En)
	assert.Equal(t, "洛杉矶", n2.Zh)

	n1, n2 = ResolveAdminNames(a1, a2, "DE", "BR", "")
	assert.Equal(t, "Brandenburg", n1.En)
	assert.Equal(t, AdminEntry{}, n2)

	n1, n2 = ResolveAdminNames(a1, a2, "XX", "YY", "ZZ")
	assert.Equal(t, AdminEntry{}, n1)
	assert.Equal(t, AdminEntry{}, n2)

	n1, n2 = ResolveAdminNames(a1, a2, "", "CA", "")
	assert.Equal(t, AdminEntry{}, n1)
	assert.Equal(t, AdminEntry{}, n2)
}
