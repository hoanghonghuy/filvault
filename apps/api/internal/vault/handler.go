package vault

import (
	"net/http"
	"strings"

	"filvault/internal/apperr"
	"filvault/internal/auth"
	"filvault/internal/platform/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(g *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	vaultGroup := g.Group("/vault", middleware...)
	vaultGroup.GET("/status", h.status)
	vaultGroup.POST("/setup", h.setup)
	vaultGroup.POST("/unlock", h.unlock)
	vaultGroup.POST("/change-pin", h.changePin)
	vaultGroup.POST("/reset", h.reset)

	secured := vaultGroup.Group("", h.VaultRequired())
	secured.GET("/files", h.listFiles)
	secured.POST("/items", h.moveToVault)
	secured.POST("/items/remove", h.moveFromVault)
}

func (h *Handler) VaultRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := userIDFrom(c)
		if !ok {
			httpx.Error(c, apperr.Unauthorized)
			c.Abort()
			return
		}

		rawToken := c.GetHeader("X-Vault-Token")
		if rawToken == "" {
			httpx.Error(c, apperr.New(http.StatusUnauthorized, "VAULT_LOCKED", "Vault is locked"))
			c.Abort()
			return
		}

		sub, err := h.svc.ValidateToken(strings.TrimSpace(rawToken))
		if err != nil || sub != userID {
			httpx.Error(c, apperr.New(http.StatusUnauthorized, "VAULT_LOCKED", "Vault is locked"))
			c.Abort()
			return
		}

		c.Next()
	}
}

func userIDFrom(c *gin.Context) (string, bool) {
	v, ok := c.Get(auth.CtxUserID)
	if !ok {
		return "", false
	}
	id, _ := v.(string)
	return id, id != ""
}

func (h *Handler) status(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}

	rawToken := c.GetHeader("X-Vault-Token")
	st, err := h.svc.Status(c.Request.Context(), userID, rawToken)
	if err != nil {
		httpx.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, st)
}

type setupReq struct {
	Pin string `json:"pin"`
}

func (h *Handler) setup(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}

	var req setupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}

	session, err := h.svc.Setup(c.Request.Context(), userID, req.Pin)
	if err != nil {
		httpx.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, session)
}

type unlockReq struct {
	Pin string `json:"pin"`
}

func (h *Handler) unlock(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}

	var req unlockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}

	session, err := h.svc.Unlock(c.Request.Context(), userID, req.Pin)
	if err != nil {
		httpx.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, session)
}

type changePinReq struct {
	CurrentPin string `json:"currentPin"`
	NewPin     string `json:"newPin"`
}

func (h *Handler) changePin(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}

	var req changePinReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}

	if err := h.svc.ChangePin(c.Request.Context(), userID, req.CurrentPin, req.NewPin); err != nil {
		httpx.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

type resetReq struct {
	AccountPassword string `json:"accountPassword"`
	NewPin          string `json:"newPin"`
}

func (h *Handler) reset(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}

	var req resetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}

	session, err := h.svc.ResetPin(c.Request.Context(), userID, req.AccountPassword, req.NewPin)
	if err != nil {
		httpx.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *Handler) listFiles(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}

	files, err := h.svc.ListFiles(c.Request.Context(), userID)
	if err != nil {
		httpx.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

type itemsReq struct {
	FileIDs []string `json:"fileIds"`
}

func (h *Handler) moveToVault(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}

	var req itemsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}

	if err := h.svc.MoveToVault(c.Request.Context(), userID, req.FileIDs); err != nil {
		httpx.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "count": len(req.FileIDs)})
}

func (h *Handler) moveFromVault(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		httpx.Error(c, apperr.Unauthorized)
		return
	}

	var req itemsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Validation(c)
		return
	}

	if err := h.svc.MoveFromVault(c.Request.Context(), userID, req.FileIDs); err != nil {
		httpx.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "count": len(req.FileIDs)})
}
