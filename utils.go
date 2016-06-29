package ibis

import (
	"fmt"
	"net/http"

	"github.com/dmajkic/ibis/jsonapi"

	"github.com/gin-gonic/gin"
)

// Resources is a helper function to set jsonapi routes for model
func (p *Ibis) Resources(router *gin.RouterGroup, name, parent string, model interface{}) {

	if meta, ok := model.(jsonapi.MetaFiller); ok {
		router.GET("/"+name+"/:id", p.getIdMetaHandler(model, meta))
	} else {
		router.GET("/"+name+"/:id", p.getIdHandler(model))
	}

	router.GET("/"+name, p.getHandler(model, parent))
	router.DELETE("/"+name+"/:id", p.deleteHandler(model))
	router.PATCH("/"+name+"/:id", p.patchHandler(model))
	router.POST("/"+name, p.postHandler(model))
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
func (p *Ibis) getHandler(model interface{}, parent string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var parent_id string

		if len(parent) == 0 {
			parent_id = c.MustGet("user_id").(string)
		} else {
			parent_id = c.DefaultQuery(parent, "")
		}

		result, err := p.Db.FindAll(model, parent_id, c.Request.URL.RawQuery)
		if err != nil {
			JSONError(c, http.StatusInternalServerError, err)
			return
		}

		c.JSON(200, result)
	}
}

// IdMetaHandler is for single model with support for MetaFiller interface
func (p *Ibis) getIdMetaHandler(model interface{}, meta jsonapi.MetaFiller) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		result, err := p.Db.FindRecord(model, id, c.Request.URL.RawQuery)

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

		meta.AddMeta(p.Db, result)

		c.JSON(http.StatusOK, result)
	}

}

// Handler to return single JSONAPI resource for specified id
func (p *Ibis) getIdHandler(model interface{}) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		result, err := p.Db.FindRecord(model, id, c.Request.URL.RawQuery)

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
func (p *Ibis) deleteHandler(model interface{}) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		err := p.Db.Delete(model, id)

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
func (p *Ibis) patchHandler(model interface{}) func(c *gin.Context) {
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

		if err := p.Db.Update(model, id, data); err != nil {
			JSONError(c, 422, err)
			return
		}

		c.JSON(http.StatusAccepted, model)
	}
}

// Handler for POST to create JSONAPI resource
func (p *Ibis) postHandler(model interface{}) func(c *gin.Context) {
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

		if result, err = p.Db.Create(model, data); err != nil {
			JSONError(c, http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}
