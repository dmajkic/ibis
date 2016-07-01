package ibis

import (
	"fmt"
	"net/http"

	"github.com/dmajkic/ibis/jsonapi"

	"github.com/gin-gonic/gin"
)

// Resources is a helper function to set jsonapi routes for model
func (s *Server) Resources(router *gin.RouterGroup, name, parent string, model interface{}) {

	if meta, ok := model.(jsonapi.MetaFiller); ok {
		router.GET("/"+name+"/:id", s.getIdMetaHandler(model, meta))
	} else {
		router.GET("/"+name+"/:id", s.getIdHandler(model))
	}

	router.GET("/"+name, s.getHandler(model, parent))
	router.DELETE("/"+name+"/:id", s.deleteHandler(model))
	router.PATCH("/"+name+"/:id", s.patchHandler(model))
	router.POST("/"+name, s.postHandler(model))
	// OPTIONS supported via CORSMiddleware()
}

// Helper fuction to return JSONAPI error with errorcode
func JSONError(c *gin.Context, errorCode int, err error) {
	c.JSON(errorCode, jsonapi.DocError(errorCode, err))
}

// Helper function to return JSONAPI 500 Internal Server
func JSONError500(c *gin.Context, err error) {
	c.JSON(500, jsonapi.DocError(500, err))
}

// Helper function to return JSONAPI 422 response
func JSONError422(c *gin.Context, source string, err error) {
	data := jsonapi.DocError(422, err)
	data.Errors[0].Source.Pointer = source
	c.JSON(422, data)
}

// Handler to return JSONAPI resource array, with optional parent
func (s *Server) getHandler(model interface{}, parent string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var parent_id string

		if len(parent) == 0 {
			parent_id = c.MustGet("user_id").(string)
		} else {
			parent_id = c.DefaultQuery(parent, "")
		}

		result, err := s.Db.FindAll(model, parent_id, c.Request.URL.RawQuery)
		if err != nil {
			JSONError(c, http.StatusInternalServerError, err)
			return
		}

		c.JSON(200, result)
	}
}

// IdMetaHandler is for single model with support for MetaFiller interface
func (s *Server) getIdMetaHandler(model interface{}, meta jsonapi.MetaFiller) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		result, err := s.Db.FindRecord(model, id, c.Request.URL.RawQuery)

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

		meta.AddMeta(s.Db, result)

		c.JSON(http.StatusOK, result)
	}

}

// Handler to return single JSONAPI resource for specified id
func (s *Server) getIdHandler(model interface{}) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		result, err := s.Db.FindRecord(model, id, c.Request.URL.RawQuery)

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
func (s *Server) deleteHandler(model interface{}) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		err := s.Db.Delete(model, id)

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
func (s *Server) patchHandler(model interface{}) func(c *gin.Context) {
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
		if id != data.Data.Id {
			JSONError(c, 422, fmt.Errorf("Wrong resource for update"))
			return
		}

		if err := s.Db.Update(model, id, data); err != nil {
			JSONError(c, 422, err)
			return
		}

		c.JSON(http.StatusAccepted, model)
	}
}

// Handler for POST to create JSONAPI resource
func (s *Server) postHandler(model interface{}) func(c *gin.Context) {
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

		if result, err = s.Db.Create(model, data); err != nil {
			JSONError(c, http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}
