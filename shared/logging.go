package shared

type Logger interface {
	func(message any, other ...any)
	func(message error, other ...error)
}
