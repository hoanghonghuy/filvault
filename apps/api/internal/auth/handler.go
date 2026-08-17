package auth

import (
	"filnest/internal/apperr"
	"filnest/internal/platform/httpx"
	"filnest/internal/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/auth/register", h.register)
	g.POST("/auth/login", h.login)
	g.POST("/auth/refresh", h.refresh)
	g.POST("/auth/logout", h.AuthRequired(), h.logout)
	g.POST("/auth/verify-email", h.verifyEmail)
	g.POST("/auth/resend-verification", h.resendVerification)
	g.GET("/users/me", h.AuthRequired(), h.me)
}

func (h *Handler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := bearer(c.GetHeader("Authorization"))
		if err != nil {
			c.Abort()
			httpx.Error(c, err)
			return
		}
		userID, err := h.svc.ParseAccess(raw)
		if err != nil {
			c.Abort()
			httpx.Error(c, err)
			return
		}
		c.Set(ctxUserID, userID)
		c.Next()
	}
}

const ctxUserID = "authUserID"

func userIDFrom(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

type registerReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	InviteCode  string `json:"inviteCode"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshReq struct {
	RefreshToken string `json:"refreshToken"`
}

type verifyEmailReq struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type resendVerificationReq struct {
	Email string `json:"email"`
}

func (h *Handler) register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	session, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.DisplayName, req.InviteCode)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, sessionJSON(session))
}

func (h *Handler) login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	session, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, sessionJSON(session))
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	session, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, sessionJSON(session))
}

func (h *Handler) logout(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	if err := h.svc.Logout(c.Request.Context(), userID, req.RefreshToken); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) me(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}
	u, err := h.svc.Me(c.Request.Context(), userID)
	if err != nil {
		httpx.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, user.PublicFrom(u))
}

func (h *Handler) verifyEmail(c *gin.Context) {
	var req verifyEmailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	if err := h.svc.VerifyEmail(c.Request.Context(), req.Email, req.Code); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) resendVerification(c *gin.Context) {
	var req resendVerificationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}
	if err := h.svc.ResendVerification(c.Request.Context(), req.Email); err != nil {
		httpx.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func sessionJSON(s Session) gin.H {
	return gin.H{
		"user":         user.PublicFrom(s.User),
		"accessToken":  s.AccessToken,
		"refreshToken": s.RefreshToken,
	}
}
