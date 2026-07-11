package api

import (
	"net/http"
	"time"

	"openshield-manager/internal/db"
	"openshield-manager/internal/models"
	"openshield-manager/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getOrgScope returns the org ID from context and applies TenantScope.
// Returns nil orgID for super admins (no scoping).
func getOrgScope(c *gin.Context) *uuid.UUID {
	orgID, _ := utils.GetOrgID(c)
	return orgID
}

// GetGroups returns all agent groups scoped to the user's organization.
func GetGroups(c *gin.Context) {
	orgID := getOrgScope(c)
	var groups []models.AgentGroup
	if err := db.DB.Scopes(db.TenantScope(orgID)).Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// CreateGroup creates a new agent group scoped to the user's organization.
func CreateGroup(c *gin.Context) {
	var group models.AgentGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Assign org from authenticated user
	orgID, _ := utils.GetOrgID(c)
	group.OrganizationID = orgID

	if err := db.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	c.JSON(http.StatusCreated, group)
}

// GetGroup returns a specific group by ID, scoped to the user's organization.
func GetGroup(c *gin.Context) {
	id := c.Param("id")
	orgID := getOrgScope(c)

	var group models.AgentGroup
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", id).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// For static groups, get member agents
	if !group.IsDynamic {
		var memberships []models.GroupMembership
		db.DB.Scopes(db.TenantScope(orgID)).Where("group_id = ?", id).Find(&memberships)

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

	// For dynamic groups, evaluate criteria, scoped to org
	var agents []models.Agent
	query := db.DB.Scopes(db.TenantScope(orgID))

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

// UpdateGroup updates a group, scoped to the user's organization.
func UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	orgID := getOrgScope(c)

	var group models.AgentGroup
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", id).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var updates models.AgentGroup
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preserve existing org assignment
	updates.OrganizationID = group.OrganizationID

	if err := db.DB.Model(&group).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update group"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// DeleteGroup deletes a group and its memberships, scoped to the user's organization.
func DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	orgID := getOrgScope(c)

	var group models.AgentGroup
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", id).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	// Delete memberships first
	db.DB.Scopes(db.TenantScope(orgID)).Where("group_id = ?", id).Delete(&models.GroupMembership{})

	// Delete the group
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", id).Delete(&models.AgentGroup{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete group"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted"})
}

// AddAgentsToGroup adds agents to a static group, scoped to the user's organization.
func AddAgentsToGroup(c *gin.Context) {
	groupID := c.Param("id")
	orgID := getOrgScope(c)

	var group models.AgentGroup
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", groupID).First(&group).Error; err != nil {
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
			GroupID:        groupUUID,
			AgentID:        agentUUID,
			OrganizationID: orgID,
		}
		db.DB.Create(&membership)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agents added to group"})
}

// RemoveAgentsFromGroup removes agents from a static group, scoped to the user's organization.
func RemoveAgentsFromGroup(c *gin.Context) {
	groupID := c.Param("id")
	orgID := getOrgScope(c)

	var group models.AgentGroup
	if err := db.DB.Scopes(db.TenantScope(orgID)).Where("id = ?", groupID).First(&group).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	var req struct {
		AgentIDs []string `json:"agent_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, agentID := range req.AgentIDs {
		db.DB.Scopes(db.TenantScope(orgID)).
			Where("group_id = ? AND agent_id = ?", groupID, agentID).
			Delete(&models.GroupMembership{})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agents removed from group"})
}
