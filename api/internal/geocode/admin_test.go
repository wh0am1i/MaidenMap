package geocode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveAdminNames(t *testing.T) {
	a1 := map[string]string{"US.CA": "California", "DE.BR": "Brandenburg"}
	a2 := map[string]string{"US.CA.037": "Los Angeles"}

	n1, n2 := ResolveAdminNames(a1, a2, "US", "CA", "037")
	assert.Equal(t, "California", n1)
	assert.Equal(t, "Los Angeles", n2)

	n1, n2 = ResolveAdminNames(a1, a2, "DE", "BR", "")
	assert.Equal(t, "Brandenburg", n1)
	assert.Equal(t, "", n2)

	n1, n2 = ResolveAdminNames(a1, a2, "XX", "YY", "ZZ")
	assert.Equal(t, "", n1)
	assert.Equal(t, "", n2)

	// Empty country should return empty (no lookup)
	n1, n2 = ResolveAdminNames(a1, a2, "", "CA", "")
	assert.Equal(t, "", n1)
	assert.Equal(t, "", n2)
}
