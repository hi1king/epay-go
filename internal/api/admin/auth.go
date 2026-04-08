// internal/api/admin/auth.go
package admin

import (
	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/jwt"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 管理员登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}

	adminService := service.NewAdminUserService()
	admin, err := adminService.Login(&service.AdminUserLoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		response.Error(c, response.CodeUnauthorized, err.Error())
		return
	}

	// 生成 Token
	token, err := jwt.GenerateToken(admin.ID, admin.Username, jwt.TokenTypeAdmin)
	if err != nil {
		response.ServerError(c, "生成Token失败")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"admin": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
			"role":     admin.Role,
		},
	})
}

// Logout 退出登录
func Logout(c *gin.Context) {
	// JWT 无状态，客户端删除 token 即可
	response.Success(c, nil)
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
}

func UpdatePassword(c *gin.Context) {
	adminID := middleware.GetUserID(c)

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	adminService := service.NewAdminUserService()
	if err := adminService.UpdatePassword(adminID, req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, response.CodeParamError, err.Error())
		return
	}

	response.Success(c, nil)
}
