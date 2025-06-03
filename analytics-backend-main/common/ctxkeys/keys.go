package ctxkeys

type contextKey string

const (
	TraceIDKey   contextKey = "trace_id"   // OpenTelemetry trace ID
	RequestIDKey contextKey = "request_id" // HTTP Request ID
	UserIDKey    contextKey = "user_id"    // ID пользователя из JWT
	RolesKey     contextKey = "roles"      // Роли пользователя
	IPAddressKey contextKey = "ip_address" // IP-адрес клиента
	UserAgentKey contextKey = "user_agent" // User-Agent клиента
	JTI          contextKey = "jti"        // JWT ID (уникальный идентификатор токена)
	LoggerKey    contextKey = "logger"     // Контекстный логгер (опционально)
)
