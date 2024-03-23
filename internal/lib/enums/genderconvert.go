package enums

const (
	genderMale   = "Male"
	genderFemale = "Female"
	dbMale       = "m"
	dbFemale     = "f"
)

func GenderConvertFromString(gender string) string {
	if gender == genderMale {
		return dbMale
	} else if gender == genderFemale {
		return dbFemale
	} else {
		return ""
	}
}
