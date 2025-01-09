package sanitize

import "regexp"

var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&’*+/=?^_{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*.[a-zA-Z]+$`)
var CatalogueRegex = regexp.MustCompile(`\+8869[0-9]{8}`)
