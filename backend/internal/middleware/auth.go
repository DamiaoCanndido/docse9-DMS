package middleware

import (
	"strings"

	"github.com/DamiaoCanndido/docse9-DMS/backend/internal/domain"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/response"
	"github.com/DamiaoCanndido/docse9-DMS/backend/pkg/security"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware intercepta requisições protegidas e valida o token JWT.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "cabeçalho de autorização ausente")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "formato do cabeçalho de autorização inválido")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := security.ValidateToken(tokenStr)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		// Injeta as informações do usuário autenticado no contexto
		c.Set("user", claims)

		// Se o usuário precisa trocar a senha, impede outros acessos
		if claims.MustChangePassword {
			if !(c.Request.Method == "POST" && strings.HasSuffix(c.Request.URL.Path, "/users/me/change-password")) {
				response.Forbidden(c, "troca de senha obrigatória")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireRole exige que o usuário autenticado possua uma das roles permitidas.
func RequireRole(allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("user")
		if !exists {
			response.Unauthorized(c, "usuário não autenticado")
			c.Abort()
			return
		}

		claims, ok := userVal.(*security.UserClaims)
		if !ok {
			response.InternalError(c)
			c.Abort()
			return
		}

		userRole := domain.Role(claims.Role)
		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "permissão insuficiente para acessar este recurso")
		c.Abort()
	}
}
