package geocode

// AdminEntry is a bilingual administrative division name.
// JSON tags let it round-trip through the admin_codes.json file written by
// update-data.
type AdminEntry struct {
	En string `json:"en"`
	Zh string `json:"zh"`
}

// ResolveAdminNames looks up admin1/admin2 display names from code maps.
// Empty inputs or missing entries return zero-value AdminEntry (not an error).
func ResolveAdminNames(a1, a2 map[string]AdminEntry, country, admin1Code, admin2Code string) (AdminEntry, AdminEntry) {
	if country == "" {
		return AdminEntry{}, AdminEntry{}
	}
	var n1, n2 AdminEntry
	if admin1Code != "" {
		n1 = a1[country+"."+admin1Code]
	}
	if admin1Code != "" && admin2Code != "" {
		n2 = a2[country+"."+admin1Code+"."+admin2Code]
	}
	return n1, n2
}
