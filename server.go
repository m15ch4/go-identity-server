package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginUserBody struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type CreateUserBody struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Role      string `json:"role"`
	Age       int    `json:"age"`
}

type CreateVMBody struct {
	Name     string `json:"name"`
	NumCPUs  int    `json:"numcpus"`
	MemoryMB int    `json:"memorymb"`
}

type Server struct {
	tokenService TokenService
	userService  UserService
	vmService    VMService
	port         int
	host         string
}

// NewServer creates a new server with the given token service, users and port.
func NewServer(tokenService TokenService, userService UserService, vmService VMService, port int, host string) *Server {
	return &Server{tokenService: tokenService, userService: userService, vmService: vmService, port: port, host: host}
}

// Run configures http routing using gin library and starts the server.
func (s *Server) Run() {
	r := gin.Default()
	r.Use(gin.Recovery())
	r.POST("/login", s.login)
	authorized := r.Group("/")
	authorized.Use(s.AuthMiddleware())
	{
		authorized.GET("/", s.home)
		authorized.POST("/users", s.createUser)
		authorized.GET("/users", s.listUsers)
		authorized.GET("/users/:id", s.getUser)
		authorized.PUT("/users/:id", s.updateUser)
		authorized.DELETE("/users/:id", s.deleteUser)
		authorized.POST("/reject", s.reject)
		authorized.GET("/rejected", s.listRejected)
		authorized.POST("/vms", s.createVM)
	}

	r.Run(fmt.Sprintf(":%d", s.port))
	fmt.Printf("Server listening on port %d\n", s.port)
}

// AuthMiddleware returns a gin.HandlerFunc that checks if the user is authenticated.
func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get authorization token from header
		tokenString := ctx.Request.Header.Get("Authorization")
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": "No token provided",
			})
			ctx.Abort()
			return
		}

		// strip "Bearer " from token string
		tokenString = tokenString[7:]

		// validate token
		user, err := s.tokenService.ValidateToken(tokenString)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			ctx.Abort()
			return
		}

		// set user in context
		ctx.Set("user", user)

		ctx.Next()
	}
}

func (s *Server) home(ctx *gin.Context) {
	// get user from context
	user := ctx.MustGet("user").(User)

	// return welcome message
	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Hello %s", user.Name),
	})
}

// createUser creates a new user with the given name and password.
func (s *Server) createUser(ctx *gin.Context) {
	var createUserBody CreateUserBody
	if err := ctx.ShouldBindJSON(&createUserBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// create new user
	user, err := s.userService.CreateUser(&createUserBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// return new user info
	ctx.JSON(http.StatusCreated, user)
}

// getUser is a handler that returns the user with the given ID.
func (s *Server) getUser(ctx *gin.Context) {
	// get user based on id
	id := ctx.Param("id")

	user, err := s.userService.GetUser(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// deleteUser is a handler that deletes the user with the given ID.
func (s *Server) deleteUser(ctx *gin.Context) {
	// get user id from path
	id := ctx.Param("id")

	//remove user with given id from users list
	if err := s.userService.DeleteUser(id); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// updateUser updates the user with the given id
// using the given user body.
func (s *Server) updateUser(ctx *gin.Context) {
	var updateUserBody CreateUserBody
	if err := ctx.ShouldBindJSON(&updateUserBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	id := ctx.Param("id")

	user, err := s.userService.UpdateUser(id, &updateUserBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, user)
}

// listUsers is a handler that returns a list of all users.
func (s *Server) listUsers(ctx *gin.Context) {
	users, err := s.userService.ListUsers()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, users)
}

// login is a handler that authenticates the given user and returns a JWT token.
func (s *Server) login(ctx *gin.Context) {
	// validate loginuser password and generate token
	var loginUserBody LoginUserBody
	if err := ctx.ShouldBindJSON(&loginUserBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	user, err := s.userService.ValidateCredentials(&loginUserBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	token, err := s.tokenService.GenerateToken(user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// reject is a handler that rejects the token
func (s *Server) reject(ctx *gin.Context) {
	tokenString := ctx.Request.Header.Get("Authorization")
	if tokenString == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "No token provided",
		})
		return
	}
	// strip "Bearer " from token string
	tokenString = tokenString[7:]

	err := s.tokenService.RejectToken(tokenString)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Token rejected",
	})
}

// listRejected is a handler that returns the list of rejected tokens.
func (s *Server) listRejected(ctx *gin.Context) {
	tokens, err := s.tokenService.GetRejectedTokens()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"tokens": tokens,
	})
}

func (s *Server) createVM(ctx *gin.Context) {
	var createVMBody CreateVMBody
	if err := ctx.ShouldBindJSON(&createVMBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	task, err := s.vmService.CreateVM(&createVMBody)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, task)
}
