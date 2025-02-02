package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"slashbase.com/backend/internal/controllers"
	"slashbase.com/backend/internal/middlewares"
	"slashbase.com/backend/internal/utils"
	"slashbase.com/backend/internal/views"
)

type QueryHandlers struct{}

var queryController controllers.QueryController

func (QueryHandlers) RunQuery(c *gin.Context) {
	var runBody struct {
		DBConnectionID string `json:"dbConnectionId"`
		Query          string `json:"query"`
	}
	c.BindJSON(&runBody)
	authUser := middlewares.GetAuthUser(c)

	data, err := queryController.RunQuery(authUser, runBody.DBConnectionID, runBody.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (QueryHandlers) GetData(c *gin.Context) {
	dbConnId := c.Param("dbConnId")

	schema := c.Query("schema")
	name := c.Query("name")
	fetchCount := c.Query("count") == "true"
	limitStr := c.Query("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 0
	}
	offsetStr := c.Query("offset")
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		offset = int64(0)
	}
	filter, _ := c.GetQueryArray("filter[]")
	sort, _ := c.GetQueryArray("sort[]")
	authUser := middlewares.GetAuthUser(c)
	authUserProjectIds := middlewares.GetAuthUserProjectIds(c)

	data, err := queryController.GetData(authUser, authUserProjectIds, dbConnId, schema, name, fetchCount, limit, offset, filter, sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (QueryHandlers) GetDataModels(c *gin.Context) {
	dbConnId := c.Param("dbConnId")
	authUser := middlewares.GetAuthUser(c)
	authUserProjectIds := middlewares.GetAuthUserProjectIds(c)

	dataModels, err := queryController.GetDataModels(authUser, authUserProjectIds, dbConnId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dataModels,
	})
}

func (QueryHandlers) GetSingleDataModel(c *gin.Context) {
	dbConnId := c.Param("dbConnId")

	schema := c.Query("schema")
	name := c.Query("name")
	authUser := middlewares.GetAuthUser(c)
	authUserProjectIds := middlewares.GetAuthUserProjectIds(c)

	data, err := queryController.GetSingleDataModel(authUser, authUserProjectIds, dbConnId, schema, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (QueryHandlers) AddSingleDataModelField(c *gin.Context) {
	var reqBody struct {
		DBConnectionID string `json:"dbConnectionId"`
		Schema         string `json:"schema"`
		Name           string `json:"name"`
		FieldName      string `json:"fieldName"`
		DataType       string `json:"dataType"`
	}
	c.BindJSON(&reqBody)
	authUser := middlewares.GetAuthUser(c)
	authUserProjectIds := middlewares.GetAuthUserProjectIds(c)

	data, err := queryController.AddSingleDataModelField(authUser, authUserProjectIds, reqBody.DBConnectionID, reqBody.Schema, reqBody.Name, reqBody.FieldName, reqBody.DataType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (QueryHandlers) DeleteSingleDataModelField(c *gin.Context) {
	var reqBody struct {
		DBConnectionID string `json:"dbConnectionId"`
		Schema         string `json:"schema"`
		Name           string `json:"name"`
		FieldName      string `json:"fieldName"`
	}
	c.BindJSON(&reqBody)
	authUser := middlewares.GetAuthUser(c)
	authUserProjectIds := middlewares.GetAuthUserProjectIds(c)

	data, err := queryController.DeleteSingleDataModelField(authUser, authUserProjectIds, reqBody.DBConnectionID, reqBody.Schema, reqBody.Name, reqBody.FieldName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (QueryHandlers) AddData(c *gin.Context) {
	dbConnId := c.Param("dbConnId")
	var addBody struct {
		Schema string                 `json:"schema"`
		Name   string                 `json:"name"`
		Data   map[string]interface{} `json:"data"`
	}
	c.BindJSON(&addBody)
	authUser := middlewares.GetAuthUser(c)

	data, err := queryController.AddData(authUser, dbConnId, addBody.Schema, addBody.Name, addBody.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (QueryHandlers) DeleteData(c *gin.Context) {
	dbConnId := c.Param("dbConnId")
	authUser := middlewares.GetAuthUser(c)
	var deleteBody struct {
		Schema string   `json:"schema"`
		Name   string   `json:"name"`
		IDs    []string `json:"ids"` // ctid for postgres, _id for mongo
	}
	c.BindJSON(&deleteBody)

	data, err := queryController.DeleteData(authUser, dbConnId, deleteBody.Schema, deleteBody.Name, deleteBody.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (QueryHandlers) UpdateSingleData(c *gin.Context) {
	dbConnId := c.Param("dbConnId")
	authUser := middlewares.GetAuthUser(c)
	var updateBody struct {
		Schema     string `json:"schema"`
		Name       string `json:"name"`
		ID         string `json:"id"`
		ColumnName string `json:"columnName"`
		Value      string `json:"value"`
	}
	c.BindJSON(&updateBody)

	data, err := queryController.UpdateSingleData(authUser, dbConnId, updateBody.Schema, updateBody.Name, updateBody.ID, updateBody.ColumnName, updateBody.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (QueryHandlers) SaveDBQuery(c *gin.Context) {
	dbConnId := c.Param("dbConnId")
	authUser := middlewares.GetAuthUser(c)
	authUserProjectIds := middlewares.GetAuthUserProjectIds(c)
	var createBody struct {
		Name    string `json:"name"`
		Query   string `json:"query"`
		QueryID string `json:"queryId"`
	}
	c.BindJSON(&createBody)

	queryObj, err := queryController.SaveDBQuery(authUser, authUserProjectIds, dbConnId, createBody.Name, createBody.Query, createBody.QueryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    views.BuildDBQueryView(queryObj),
	})
}

func (QueryHandlers) GetDBQueriesInDBConnection(c *gin.Context) {
	dbConnID := c.Param("dbConnId")
	authUserProjectIds := middlewares.GetAuthUserProjectIds(c)

	dbQueries, err := queryController.GetDBQueriesInDBConnection(authUserProjectIds, dbConnID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	dbQueryViews := []views.DBQueryView{}
	for _, dbQuery := range dbQueries {
		dbQueryViews = append(dbQueryViews, *views.BuildDBQueryView(dbQuery))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dbQueryViews,
	})
}

func (QueryHandlers) GetSingleDBQuery(c *gin.Context) {
	queryID := c.Param("queryId")
	authUserProjectIds := middlewares.GetAuthUserProjectIds(c)

	dbQuery, err := queryController.GetSingleDBQuery(authUserProjectIds, queryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    views.BuildDBQueryView(dbQuery),
	})
}

func (QueryHandlers) GetQueryHistoryInDBConnection(c *gin.Context) {
	dbConnID := c.Param("dbConnId")
	authUser := middlewares.GetAuthUser(c)
	authUserProjectIds := middlewares.GetAuthUserProjectIds(c)

	beforeInt, err := strconv.ParseInt(c.Query("before"), 10, 64)
	var before time.Time
	if err != nil {
		before = time.Now()
	} else {
		before = utils.UnixNanoToTime(beforeInt)
	}

	dbQueryLogs, next, err := queryController.GetQueryHistoryInDBConnection(authUser, authUserProjectIds, dbConnID, before)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	dbQueryLogViews := []views.DBQueryLogView{}
	for _, dbQueryLog := range dbQueryLogs {
		dbQueryLogViews = append(dbQueryLogViews, *views.BuildDBQueryLogView(dbQueryLog))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"list": dbQueryLogViews,
			"next": next,
		},
	})
}
