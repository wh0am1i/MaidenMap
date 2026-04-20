package geocode

// ResolveAdminNames looks up admin1/admin2 display names from code maps.
// Empty inputs or missing entries return empty strings (not an error).
func ResolveAdminNames(a1, a2 map[string]string, country, admin1Code, admin2Code string) (string, string) {
	if country == "" {
		return "", ""
	}
	var n1, n2 string
	if admin1Code != "" {
		n1 = a1[country+"."+admin1Code]
	}
	if admin1Code != "" && admin2Code != "" {
		n2 = a2[country+"."+admin1Code+"."+admin2Code]
	}
	return n1, n2
}
