package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/golang-jwt/jwt"
	"github.com/robwittman/chushi/internal/resource/agent"
	"github.com/robwittman/chushi/internal/resource/run"
	"github.com/robwittman/chushi/internal/resource/workspaces"
	"github.com/robwittman/chushi/internal/server/helpers"
	"github.com/robwittman/chushi/pkg/types"
	"net/http"
	"os"
	"strings"
)

type AgentsController struct {
	Repository          agent.AgentRepository
	RunsRepository      run.RunRepository
	WorkspaceRepository workspaces.WorkspacesRepository
}

func (ctrl *AgentsController) List(c *gin.Context) {
	orgId, err := helpers.GetOrganizationId(c)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}
	agents, err := ctrl.Repository.List(orgId)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"agents": agents,
		})
	}
}

func (ctrl *AgentsController) Get(c *gin.Context) {
	orgId, err := helpers.GetOrganizationId(c)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}
	ag, err := ctrl.Repository.FindById(orgId, c.Param("agent"))
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"agent": ag,
		})
	}
}

type CreateAgentRequest struct {
	Name string `json:"name"`
}

func (ctrl *AgentsController) Create(c *gin.Context) {
	orgId, err := helpers.GetOrganizationId(c)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	var params CreateAgentRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	ag := &agent.Agent{
		OrganizationID: orgId,
		Name:           params.Name,
	}
	if _, err := ctrl.Repository.Create(ag); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"agent": ag,
		})
	}
}

func (ctrl *AgentsController) Update(c *gin.Context) {

}

func (ctrl *AgentsController) Delete(c *gin.Context) {

}

func (ctrl *AgentsController) GetRuns(c *gin.Context) {

	params := &run.RunListParams{
		AgentId: c.Param("agent"),
	}
	if status := c.Query("status"); status != "" {
		runStatus, _ := types.ToRunStatus(status)
		params.Status = runStatus
	}
	runs, err := ctrl.RunsRepository.List(params)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"runs": runs,
	})
}

func (ctrl *AgentsController) AgentAccess(c *gin.Context) {
	input := c.Request.Header.Get("Authorization")
	if input == "" {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	input, _ = strings.CutPrefix(input, "Bearer ")

	token, err := jwt.ParseWithClaims(input, &generates.JWTAccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("parse error")
		}
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	claims := token.Claims.(*generates.JWTAccessClaims)

	if _, err := ctrl.Repository.FindByClientId(claims.Audience); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}
