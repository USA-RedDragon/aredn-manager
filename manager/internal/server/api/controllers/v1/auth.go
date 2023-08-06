package v1

import (
	"fmt"
	"net"
	"net/http"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/apimodels"
	"github.com/USA-RedDragon/aredn-manager/internal/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func POSTLogin(c *gin.Context) {
	// If the IP is from the private ip ranges, reject the login. We cannot encrypt traffic over the mesh
	if net.ParseIP(c.ClientIP()).IsPrivate() {
		fmt.Println("POSTLogin: Login from private IP")
		c.JSON(http.StatusUnavailableForLegalReasons, gin.H{"error": "Cannot encrypt traffic over the mesh. Please use the site via the internet."})
		return
	}
	session := sessions.Default(c)
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Println("POSTLogin: Unable to get DB from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("POSTLogin: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	var json apimodels.AuthLogin
	err := c.ShouldBindJSON(&json)
	if err != nil {
		fmt.Printf("POSTLogin: JSON data is invalid: %v\n", err)
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

		fmt.Printf("POSTLogin: User found %v\n", user)

		fmt.Printf("POSTLogin: %v\n", utils.HashPassword(json.Password, config.PasswordSalt))

		verified, err := utils.VerifyPassword(json.Password, user.Password, config.PasswordSalt)
		fmt.Printf("POSTLogin: Password verified %v\n", verified)
		if verified && err == nil {
			session.Set("user_id", user.ID)
			err = session.Save()
			if err != nil {
				fmt.Printf("POSTLogin: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving session"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Logged in"})
			return
		}
		fmt.Printf("POSTLogin: %v\n", err)
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
}

func GETLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		fmt.Printf("GETLogout: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
