package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Создаие и проверка токенов
type JwtManager struct {
	secretKey string
}

func NewJWTManager(secretKey string) *JwtManager {
	return &JwtManager{secretKey: secretKey}
}

// Нагрузка, кооторая валяется внутри токена
type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Генерация нового JWT токена
func (m *JwtManager) GenerateToken(userID int64, role string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), //Жизнь токена: 7 дней
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// Проверка и доставание данных из токена
func (m *JwtManager) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || token.Valid {
		return nil, errors.New("Невалидный токен")
	}

	return claims, nil
}
