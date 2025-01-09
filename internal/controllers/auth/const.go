package auth

type WhitelistAccount struct {
	UserID string
	OTP    string
}

var whitelistedAccounts = map[string]*WhitelistAccount{
	"+886900000000": {
		UserID: "merchants:01J0MSQ2X1E48YW0FD1A96N9RG",
		OTP:    "123456",
	},
	"+886900000001": {
		UserID: "drivers:01J0MSSXD5FV7RG4XMV9DJDB5F",
		OTP:    "123456",
	},
}
