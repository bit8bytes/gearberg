package accounts

type contextKey string

func (contextKey) String() string {
	return string(Key)
}

// Key is the session key used to store and retrieve the authenticated account ID.
const Key contextKey = "ACCOUNT_ID_KEY"
