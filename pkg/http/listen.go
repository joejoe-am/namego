package http

func (http *Http) Listen(addr string) error {
	// Ensure the server's handler is set to the configured handler
	http.server.Handler = http.Handler()

	// Start the server and listen on the specified address
	return http.server.ListenAndServe(addr)
}
