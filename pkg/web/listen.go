package web

func (http *Http) Listen(addr string) error {
	// Start the server and listen on the specified address
	return http.server.ListenAndServe(addr)
}
