package v1

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/apimodels"
	"github.com/USA-RedDragon/aredn-manager/internal/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	gopwned "github.com/mavjs/goPwned"
	"gorm.io/gorm"
)

func GETUsers(c *gin.Context) {
	db, ok := c.MustGet("PaginatedDB").(*gorm.DB)
	if !ok {
		fmt.Println("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	cDb, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Printf("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	users, err := models.ListUsers(db)
	if err != nil {
		fmt.Printf("GETUsers: Error getting users: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting users"})
		return
	}

	total, err := models.CountUsers(cDb)
	if err != nil {
		fmt.Printf("GETUsers: Error getting user count: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting user count"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "users": users})
}

func POSTUser(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Printf("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("POSTUser: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	var json apimodels.UserRegistration
	err := c.ShouldBindJSON(&json)
	if err != nil {
		fmt.Printf("POSTUser: JSON data is invalid: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON data is invalid"})
	} else {
		isValid, errString := json.IsValidUsername()
		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": errString})
			return
		}

		// Check that password isn't a zero string
		if json.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password cannot be blank"})
			return
		}

		// Check if the username is already taken
		var user models.User
		err := db.Find(&user, "username = ?", json.Username).Error
		if err != nil {
			fmt.Printf("POSTUser: Error getting user: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting user"})
			return
		} else if user.ID != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username is already taken"})
			return
		}

		if config.HIBPAPIKey != "" {
			goPwned := gopwned.NewClient(nil, config.HIBPAPIKey)
			h := sha1.New() //#nosec G401 -- False positive, we are not using this for crypto, just HIBP
			h.Write([]byte(json.Password))
			sha1HashedPW := fmt.Sprintf("%X", h.Sum(nil))
			frange := sha1HashedPW[0:5]
			lrange := sha1HashedPW[5:40]
			karray, err := goPwned.GetPwnedPasswords(frange, false)
			if err != nil {
				// If the error message starts with "Too many requests", then tell the user to retry in one minute
				if strings.HasPrefix(err.Error(), "Too many requests") {
					c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please try again in one minute"})
					return
				}
				fmt.Printf("POSTUser: Error getting pwned passwords: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting pwned passwords"})
				return
			}
			strKArray := string(karray)
			respArray := strings.Split(strKArray, "\r\n")

			var result int64
			for _, resp := range respArray {
				strArray := strings.Split(resp, ":")
				test := strArray[0]

				count, err := strconv.ParseInt(strArray[1], 0, 32)
				if err != nil {
					fmt.Printf("POSTUser: Error parsing pwned password count: %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing pwned password count"})
					return
				}
				if test == lrange {
					result = count
				}
			}
			if result > 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Password has been reported in a data breach. Please use another one"})
				return
			}
		}

		// argon2 the password
		hashedPassword := utils.HashPassword(json.Password, config.PasswordSalt)

		user = models.User{
			Username: json.Username,
			Password: hashedPassword,
		}
		err = db.Create(&user).Error
		if err != nil {
			fmt.Printf("POSTUser: Error creating user: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User created"})
	}
}

func GETUser(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Println("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	id := c.Param("id")
	// Convert string id into uint
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}
	user, err := models.FindUserByID(db, uint(userID))
	if err != nil {
		fmt.Printf("Error finding user: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
	}
	c.JSON(http.StatusOK, user)
}

func PATCHUser(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Println("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("POSTLogin: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	id := c.Param("id")
	idInt, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}
	var json apimodels.UserPatch
	err = c.ShouldBindJSON(&json)
	if err != nil {
		fmt.Printf("PATCHUser: JSON data is invalid: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON data is invalid"})
	} else {
		user, err := models.FindUserByID(db, uint(idInt))
		if err != nil {
			fmt.Printf("Error finding user: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
			return
		}

		if json.Username != "" {
			// Check if the username is already taken
			var existingUser models.User
			err := db.Find(&existingUser, "username = ?", json.Username).Error
			if err != nil {
				fmt.Printf("Error finding user: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding user"})
				return
			} else if existingUser.ID != 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Username is already taken"})
				return
			}
			user.Username = json.Username
		}

		if json.Password != "" {
			user.Password = utils.HashPassword(json.Password, config.PasswordSalt)
		}

		err = db.Save(&user).Error
		if err != nil {
			fmt.Printf("Error updating user: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User updated"})
	}
}

func DELETEUser(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Println("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	idUint64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	exists, err := models.UserIDExists(db, uint(idUint64))
	if err != nil {
		fmt.Printf("Error checking if user exists: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking if user exists"})
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
		return
	}

	err = models.DeleteUser(db, uint(idUint64))
	if err != nil {
		fmt.Printf("Error deleting user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func GETUserSelf(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Println("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	session := sessions.Default(c)

	userID := session.Get("user_id")
	if userID == nil {
		fmt.Println("userID not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		fmt.Println("userID cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	user, err := models.FindUserByID(db, uid)
	if err != nil {
		fmt.Printf("Error finding user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding user"})
		return
	}
	c.JSON(http.StatusOK, user)
}
