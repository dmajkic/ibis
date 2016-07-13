package ibis

import (
	"fmt"
	"net/http"

	"github.com/dmajkic/ibis/jsonapi"

	"github.com/gin-gonic/gin"
)

// ResourcesFunc is a helper function to set jsonapi routes from function
func (s *Server) ResourcesFunc(router *gin.RouterGroup, name, parent string, fn func() []interface{}) {

	models := fn()

	router.GET("/"+name+"/:id", s.getIDHandler(s.ModelDb, models))
	router.GET("/"+name, s.getHandler(s.ModelDb, models, parent))
	router.DELETE("/"+name+"/:id", s.deleteHandler(s.ModelDb, models))
	router.PATCH("/"+name+"/:id", s.patchHandler(s.ModelDb, models))
	router.POST("/"+name, s.postHandler(s.ModelDb, models))
	// OPTIONS supported via CORSMiddleware()
}

// Resources is a helper function to set jsonapi routes for model slice
func (s *Server) Resources(router *gin.RouterGroup, name, parent string, models ...interface{}) {

	router.GET("/"+name+"/:id", s.getIDHandler(s.ModelDb, models))
	router.GET("/"+name, s.getHandler(s.ModelDb, models, parent))
	router.DELETE("/"+name+"/:id", s.deleteHandler(s.ModelDb, models))
	router.PATCH("/"+name+"/:id", s.patchHandler(s.ModelDb, models))
	router.POST("/"+name, s.postHandler(s.ModelDb, models))
	// OPTIONS supported via CORSMiddleware()
}

// Resource is a helper function to set jsonapi routes for model
func (s *Server) Resource(router *gin.RouterGroup, name, parent string, model interface{}) {

	if meta, ok := model.(jsonapi.MetaFiller); ok {
		router.GET("/"+name+"/:id", s.getIDMetaHandler(s.Db, model, meta))
	} else {
		router.GET("/"+name+"/:id", s.getIDHandler(s.Db, model))
	}

	router.GET("/"+name, s.getHandler(s.Db, model, parent))
	router.DELETE("/"+name+"/:id", s.deleteHandler(s.Db, model))
	router.PATCH("/"+name+"/:id", s.patchHandler(s.Db, model))
	router.POST("/"+name, s.postHandler(s.Db, model))
	// OPTIONS supported via CORSMiddleware()
}

// JSONError is a helper fuction to return JSONAPI error with errorcode
func JSONError(c *gin.Context, errorCode int, err error) {
	c.JSON(errorCode, jsonapi.DocError(errorCode, err))
}

// JSONError500 is a helper function to return JSONAPI 500 Internal Server
func JSONError500(c *gin.Context, err error) {
	c.JSON(500, jsonapi.DocError(500, err))
}

// JSONError422 is a helper function to return JSONAPI 422 response
func JSONError422(c *gin.Context, source string, err error) {
	data := jsonapi.DocError(422, err)
	data.Errors[0].Source.Pointer = source
	c.JSON(422, data)
}

// Handler to return JSONAPI resource array, with optional parent
func (s *Server) getHandler(db jsonapi.Database, model interface{}, parent string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var parentID string

		if len(parent) == 0 {
			parentID = c.MustGet("user_id").(string)
		} else {
			parentID = c.DefaultQuery(parent, "")
		}

		result, err := db.FindAll(model, parentID, c.Request.URL.RawQuery)
		if err != nil {
			JSONError(c, http.StatusInternalServerError, err)
			return
		}

		c.JSON(200, result)
	}
}

// getIDMetaHandler is for single model with support for MetaFiller interface
func (s *Server) getIDMetaHandler(db jsonapi.Database, model interface{}, meta jsonapi.MetaFiller) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		result, err := db.FindRecord(model, id, c.Request.URL.RawQuery)

		if err == jsonapi.ErrNotFound {
			JSONError(c, http.StatusNotFound, err)
			return
		} else if err != nil {
			JSONError(c, http.StatusInternalServerError, err)
			return
		}

		if len(result.Meta) == 0 {
			result.Meta = make(map[string]interface{})
		}

		// Add data json field to doc.meta.data
		if data, ok := result.Data.Attributes["data"]; ok {
			result.Meta["data"] = data
			delete(result.Data.Attributes, "data")
		}

		meta.AddMeta(db, result)

		c.JSON(http.StatusOK, result)
	}

}

// Handler to return single JSONAPI resource for specified id
func (s *Server) getIDHandler(db jsonapi.Database, model interface{}) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		result, err := db.FindRecord(model, id, c.Request.URL.RawQuery)

		if err == jsonapi.ErrNotFound {
			JSONError(c, http.StatusNotFound, err)
			return
		} else if err != nil {
			JSONError(c, http.StatusInternalServerError, err)
			return
		}

		if len(result.Meta) == 0 {
			result.Meta = make(map[string]interface{})
		}

		// Add data json field to doc.meta.data
		if data, ok := result.Data.Attributes["data"]; ok {
			result.Meta["data"] = data
			delete(result.Data.Attributes, "data")
		}

		c.JSON(http.StatusOK, result)
	}
}

// Handler to delete JSONAPI resource
func (s *Server) deleteHandler(db jsonapi.Database, model interface{}) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		err := db.Delete(model, id)

		if err == jsonapi.ErrNotFound {
			c.AbortWithStatus(http.StatusNoContent)
			return
		} else if err != nil {
			JSONError(c, http.StatusInternalServerError, err)
			return
		}

		c.AbortWithStatus(http.StatusNoContent)
	}
}

// Handler for PATCH to update JSONAPI resource
func (s *Server) patchHandler(db jsonapi.Database, model interface{}) func(c *gin.Context) {
	return func(c *gin.Context) {
		data := &jsonapi.DocItem{
			Data:  jsonapi.NewResource("", ""),
			Meta:  make(map[string]interface{}),
			Links: &jsonapi.Links{},
		}

		if err := c.BindJSON(&data); err != nil {
			JSONError(c, 422, err)
			return
		}

		id := c.Param("id")
		if id != data.Data.ID {
			JSONError(c, 422, fmt.Errorf("Wrong resource for update"))
			return
		}

		if err := db.Update(model, id, data); err != nil {
			JSONError(c, 422, err)
			return
		}

		c.JSON(http.StatusAccepted, model)
	}
}

// Handler for POST to create JSONAPI resource
func (s *Server) postHandler(db jsonapi.Database, model interface{}) func(c *gin.Context) {
	return func(c *gin.Context) {
		var err error
		var result *jsonapi.DocItem

		data := &jsonapi.DocItem{
			Data:  jsonapi.NewResource("", ""),
			Meta:  make(map[string]interface{}),
			Links: &jsonapi.Links{},
		}

		if err = c.BindJSON(&data); err != nil {
			JSONError(c, 422, err)
			return
		}

		if result, err = db.Create(model, data); err != nil {
			JSONError(c, http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}
