package organizations

import (
	"crypto/rand"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zarinit-routers/cloud-organizaions/pkg/models"
	"github.com/zarinit-routers/cloud-organizaions/pkg/storage/database"
)

type NewOrganizationRequest struct {
	Name string `json:"name" binding:"required"`
}

func NewHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req NewOrganizationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error("Failed bind JSON", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
		db := database.MustConnect()
		org := models.Organization{
			Name:      req.Name,
			CreatedAt: time.Now(),
		}
		if err := db.Create(org).Error; err != nil {
			log.Error("Failed create organization", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		c.JSON(http.StatusOK, org)
	}
}

func ListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.MustConnect()
		var orgs []models.Organization
		if err := db.Preload("Members").Find(&orgs).Error; err != nil {
			log.Error("Failed list organizations", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		c.JSON(http.StatusOK, gin.H{"organizations": orgs})
	}
}

func GetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			log.Error("Failed parse organization id", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
		db := database.MustConnect()
		var org models.Organization
		if err := db.Preload("Members").First(&org, id).Error; err != nil {
			log.Error("Failed get organization", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		c.JSON(http.StatusOK, org)
	}
}

type UpdateOrganizationRequest struct {
	ID   uuid.UUID `json:"id" binding:"required"`
	Name string    `json:"name" binding:"required"`
}

func UpdateHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateOrganizationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error("Failed bind JSON", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}
		db := database.MustConnect()
		var org models.Organization
		if err := db.First(&org, req.ID).Error; err != nil {
			log.Error("Failed get organization", "error", err)
			c.JSON(http.StatusNotFound, gin.H{})
			return
		}
		org.Name = req.Name
		if err := db.Save(&org).Error; err != nil {
			log.Error("Failed update organization", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		c.JSON(http.StatusOK, org)
	}
}

type GeneratePassphraseRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

func GeneratePassphraseHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GeneratePassphraseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error("Failed bind JSON", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}

		db := database.MustConnect()
		var org models.Organization
		if err := db.First(&org, req.ID).Error; err != nil {
			log.Error("Failed get organization", "error", err)
			c.JSON(http.StatusNotFound, gin.H{})
			return
		}

		p := generatePassphrase()
		org.Passphrase = &p

		if err := db.Save(&org).Error; err != nil {
			log.Error("Failed update organization", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}

		c.JSON(http.StatusOK, gin.H{"passphrase": p})
	}
}

func generatePassphrase() string {
	return rand.Text()
}
