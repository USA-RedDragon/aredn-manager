package v1

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/USA-RedDragon/aredn-manager/internal/bind"
	"github.com/USA-RedDragon/aredn-manager/internal/config"
	"github.com/USA-RedDragon/aredn-manager/internal/db/models"
	"github.com/USA-RedDragon/aredn-manager/internal/olsrd"
	"github.com/USA-RedDragon/aredn-manager/internal/server/api/apimodels"
	"github.com/USA-RedDragon/aredn-manager/internal/vtun"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GETMeshes(c *gin.Context) {
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
	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("GETMeshes: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}
	meshes, err := models.ListSupernodes(db)
	if err != nil {
		fmt.Printf("GETMeshes: Error getting meshes: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting meshes"})
		return
	}

	total, err := models.CountSupernodes(cDb)
	if err != nil {
		fmt.Printf("GETMeshes: Error getting mesh count: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting mesh count"})
		return
	}

	if config.Supernode {
		// We never add ourselves to the database as a supernode
		// but we do want to show up in the list to the users
		meshes = append(meshes, models.Supernode{
			MeshName: config.SupernodeZone,
			IPs:      []string{config.NodeIP},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total":  total,
		"meshes": meshes,
	})
}

func POSTMesh(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Printf("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("POSTMesh: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	var json apimodels.CreateMesh
	err := c.ShouldBindJSON(&json)
	if err != nil {
		fmt.Printf("POSTMesh: JSON data is invalid: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON data is invalid"})
	} else {
		isValid, errString := json.IsValidHostname()
		if !isValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": errString})
			return
		}

		for _, ip := range json.IPs {
			nip := net.ParseIP(ip)
			if nip == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "IP address is invalid"})
				return
			}
		}

		// Check if the mesh name is already taken
		var supernode models.Supernode
		err := db.Find(&supernode, "mesh_name = ?", json.Name).Error
		if err != nil {
			fmt.Printf("POSTMesh: Error getting mesh: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting mesh"})
			return
		} else if supernode.ID != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Mesh name is already taken"})
			return
		}

		supernode = models.Supernode{
			MeshName: json.Name,
			IPs:      json.IPs,
		}

		// Do the thing
		err = db.Create(&supernode).Error
		if err != nil {
			fmt.Printf("POSTMesh: Error creating mesh: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating mesh"})
			return
		}

		err = vtun.GenerateAndSave(config, db)
		if err != nil {
			fmt.Printf("POSTMesh: Error generating vtun config: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating vtun config"})
			return
		}

		err = olsrd.GenerateAndSave(config, db)
		if err != nil {
			fmt.Printf("POSTMesh: Error generating olsrd config: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating olsrd config"})
			return
		}

		err = vtun.Reload()
		if err != nil {
			fmt.Printf("POSTMesh: Error reloading vtun: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading vtun"})
			return
		}

		err = olsrd.Reload()
		if err != nil {
			fmt.Printf("POSTMesh: Error reloading olsrd: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading olsrd"})
			return
		}

		err = bind.GenerateAndSave(config, db)
		if err != nil {
			fmt.Printf("Error generating bind config: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating bind config"})
			return
		}

		err = bind.Restart()
		if err != nil {
			fmt.Printf("Error reloading bind: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading bind"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Mesh created"})
	}
}

func DELETEMesh(c *gin.Context) {
	db, ok := c.MustGet("DB").(*gorm.DB)
	if !ok {
		fmt.Println("DB cast failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	config, ok := c.MustGet("Config").(*config.Config)
	if !ok {
		fmt.Println("DELETEMesh: Unable to get Config from context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Try again later"})
		return
	}

	idUint64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid mesh ID"})
		return
	}

	exists, err := models.SupernodeIDExists(db, uint(idUint64))
	if err != nil {
		fmt.Printf("Error checking if mesh exists: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking if mesh exists"})
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mesh does not exist"})
		return
	}

	err = models.DeleteSupernode(db, uint(idUint64))
	if err != nil {
		fmt.Printf("Error deleting mesh: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting mesh"})
		return
	}

	err = vtun.GenerateAndSave(config, db)
	if err != nil {
		fmt.Printf("Error generating vtun config: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating vtun config"})
		return
	}

	err = olsrd.GenerateAndSave(config, db)
	if err != nil {
		fmt.Printf("Error generating olsrd config: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating olsrd config"})
		return
	}

	err = vtun.Reload()
	if err != nil {
		fmt.Printf("Error reloading vtun: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading vtun"})
		return
	}

	err = olsrd.Reload()
	if err != nil {
		fmt.Printf("Error reloading olsrd: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading olsrd"})
		return
	}

	err = bind.GenerateAndSave(config, db)
	if err != nil {
		fmt.Printf("Error generating bind config: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating bind config"})
		return
	}

	err = bind.Restart()
	if err != nil {
		fmt.Printf("Error reloading bind: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reloading bind"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mesh deleted"})
}
