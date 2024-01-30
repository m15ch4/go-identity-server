package main

// main initializes the services and starts the server on port 8080.
// it uses Gin to handle http routing and to serve the HTTP requests.
func main() {
	tokenService := NewTokenService("secret")
	userService := NewUserService()

	server := NewServer(tokenService, userService, 8090, "localhost")
	server.Run()
}
