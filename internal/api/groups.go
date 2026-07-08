package api

import (
	"net/http"
	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetGroups returns all agent groups
func GetGroups(c *gin.Context) {
	var groups []models.AgentGroup
	if err := db.DB.Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// CreateGroup creates a new agent group
func CreateGroup(c *gin.Context) {
	var group models.AgentGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// GetGroup returns a specific group by ID
func GetGroup(c *gin.Context) {
	id := c.Param("id")
	var group models.AgentGroup
	if err := db.DB.Where("id = ?", id).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// For static groups, get member agents
	if !group.IsDynamic {
		var memberships []models.GroupMembership
		db.DB.Where("group_id = ?", id).Find(&memberships)
		
		agentIDs := make([]string, len(memberships))
		for i, m := range memberships {
			agentIDs[i] = m.AgentID.String()
		}
		
		c.JSON(http.StatusOK, gin.H{
			"group":     group,
			"agent_ids": agentIDs,
		})
		return
	}

	// For dynamic groups, evaluate criteria
	var agents []models.Agent
	query := db.DB
	
	if group.Criteria != nil {
		if len(group.Criteria.State) > 0 {
			query = query.Where("state IN ?", group.Criteria.State)
		}
		if group.Criteria.LastSeenWithin > 0 {
			threshold := time.Now().Add(-time.Duration(group.Criteria.LastSeenWithin) * time.Second)
			query = query.Where("last_seen >= ?", threshold)
		}
	}
	
	query.Find(&agents)
	
	agentIDs := make([]string, len(agents))
	for i, a := range agents {
		agentIDs[i] = a.ID.String()
	}
	
	c.JSON(http.StatusOK, gin.H{
		"group":     group,
		"agent_ids": agentIDs,
	})
}

// UpdateGroup updates a group
func UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	var group models.AgentGroup
	if err := db.DB.Where("id = ?", id).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var updates models.AgentGroup
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.DB.Model(&group).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup deletes a group
func DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Where("id = ?", id).Delete(&models.AgentGroup{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}
	
	// Also delete memberships
	db.DB.Where("group_id = ?", id).Delete(&models.GroupMembership{})
	
	c.JSON(http.StatusOK, gin.H{"message": "Group deleted"})
}

// AddAgentsToGroup adds agents to a static group
func AddAgentsToGroup(c *gin.Context) {
	groupID := c.Param("id")
	
	var group models.AgentGroup
	if err := db.DB.Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}
	
	if group.IsDynamic {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add agents to dynamic group"})
		return
	}

	var req struct {
		AgentIDs []string `json:"agent_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	groupUUID, _ := uuid.Parse(groupID)
	for _, agentID := range req.AgentIDs {
		agentUUID, err := uuid.Parse(agentID)
		if err != nil {
			continue
		}
		
		membership := models.GroupMembership{
			GroupID: groupUUID,
			AgentID: agentUUID,
		}
		db.DB.Create(&membership)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agents added to group"})
}

// RemoveAgentsFromGroup removes agents from a static group
func RemoveAgentsFromGroup(c *gin.Context) {
	groupID := c.Param("id")
	
	var req struct {
		AgentIDs []string `json:"agent_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, agentID := range req.AgentIDs {
		db.DB.Where("group_id = ? AND agent_id = ?", groupID, agentID).Delete(&models.GroupMembership{})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agents removed from group"})
}
