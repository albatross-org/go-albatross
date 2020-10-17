package server

import (
	"fmt"

	"github.com/albatross-org/go-albatross/entries"
	"github.com/gin-gonic/gin"
)

// Server allows the viewing and querying (and in the future, editing) of entries over HTTP.
// It wraps a *entries.Collection, meaning "filtered" servers can be made, which only render a subset
// of a larger Albatross store.
// Servers can be started using the command line tool, running `albatross get server`.
type Server struct {
	collection *entries.Collection
	router     *gin.Engine
}

// NewServer returns a new server struct from an *entries.Collection.
func NewServer(collection *entries.Collection) *Server {
	server := &Server{
		collection: collection,
		router:     gin.Default(),
	}

	server.initRoutes()

	return server
}

// Serve begins accepting requests on the given port.
func (s *Server) Serve(port int) error {
	return s.router.Run(":" + fmt.Sprint(port))
}
