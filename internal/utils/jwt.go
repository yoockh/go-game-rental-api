package utils

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if len(auth) < 8 || auth[:7] != "Bearer " {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
			}
			
			tokenString := auth[7:]
			token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
			}
			
			c.Set("user", token)
			return next(c)
		}
	}
}

func RequireRoles(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get("user")
			if user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "No user token found")
			}
			
			token := user.(*jwt.Token)
			claims := token.Claims.(jwt.MapClaims)
			userRole, ok := claims["role"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid role in token")
			}
			
			for _, role := range allowedRoles {
				if userRole == role {
					return next(c)
				}
			}
			
			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
		}
	}
}

func GetUserIDFromToken(c echo.Context) uint {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	if id, ok := claims["user_id"].(float64); ok {
		return uint(id)
	}
	return 0
}

func GetRoleFromToken(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	if role, ok := claims["role"].(string); ok {
		return role
	}
	return ""
}