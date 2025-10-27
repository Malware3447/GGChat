package interfaces

// JwtInterface определяет интерфейс для работы с JWT
type JwtInterface interface {
	NewToken(UserId int) (string, error)
}
