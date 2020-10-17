package server

import "github.com/gin-contrib/cors"

// initRoutes sets up the required routes for the server.
func (s *Server) initRoutes() {
	s.router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"https://cdpn.io"},
	}))
	s.router.GET("/search", s.searchHandler)
}
