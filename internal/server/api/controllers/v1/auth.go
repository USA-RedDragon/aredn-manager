package v1

import (
	"log/slog"
	"net"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/apimodels"
	"github.com/USA-RedDragon/aredn-manager/internal/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

func POSTLogin(c *gin.Context) {
	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		slog.Error("POSTLogin: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	// If the IP is from the private ip ranges, reject the login. We cannot encrypt traffic over the mesh
	if net.ParseIP(c.ClientIP()).IsPrivate() && !slices.Contains(config.TrustedProxies, c.ClientIP()) {
		slog.Error("POSTLogin: Login from private IP")
		c.JSON(http.StatusUnavailableForLegalReasons, gin.H{"error": "Cannot encrypt traffic over the mesh. Please use the site via the internet."})
		return
	}

	session := sessions.Default(c)
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		slog.Error("POSTLogin: Unable to get DB from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	var json apimodels.AuthLogin
	err := c.ShouldBindJSON(&json)
	if err != nil {
		slog.Error("POSTLogin: JSON data is invalid", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON data is invalid"})
	} else {
		// Check that one of username is not blank
		if json.Username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be provided"})
			return
		}
		// Check that password isn't a zero string
		if json.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password cannot be blank"})
			return
		}
		var user models.User
		db.Find(&user, "username = ?", json.Username)
		if db.Error != nil {
			slog.Error("POSTLogin: Error finding user", "error", db.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
			return
		}

		slog.Debug("POSTLogin: User found", "user", user)

		verified, err := utils.VerifyPassword(json.Password, user.Password, config.PasswordSalt)
		if verified && err == nil {
			session.Set("user_id", user.ID)
			err = session.Save()
			if err != nil {
				slog.Error("POSTLogin: Error saving session", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving session"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Logged in"})
			return
		}
		slog.Error("POSTLogin: Invalid username or password", "username", json.Username)
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
}

func GETLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		slog.Error("GETLogout: Error saving session", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
