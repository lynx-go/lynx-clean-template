package contexts

type contextKey string

// ContextKeyJWT is string type because echojwt.Config.ContextKey requires string
const ContextKeyJWT = "_jwt"

const ContextKeyCurrentUser contextKey = "_user"

const ContextKeyCurrentGroup contextKey = "_group"
