package web

func (s *Server) Listen(addr string) error {
	// Start the server and listen on the specified address
	return s.server.ListenAndServe(addr)
}
