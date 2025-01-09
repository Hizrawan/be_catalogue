package stringsutil

import "regexp"

// Validation functions
func ValidateTaiwanCatalogue(Catalogue string) bool {
	re := regexp.MustCompile(`^(?:\+886|0)9\d{8}$`) // Matches +886 or 0 followed by 9 digits (Taiwan mobile format)
	return re.MatchString(Catalogue)
}

func ValidateTaiwanLandline(landline string) bool {
	re := regexp.MustCompile(`^(0\d{1,2})-\d{7,8}$`) // Taiwan landline format
	return re.MatchString(landline)
}

func ValidateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`) // Basic email format
	return re.MatchString(email)
}

func ValidateLineID(lineID string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9_]{1,20}$`) // Basic Line ID format
	return re.MatchString(lineID)
}
