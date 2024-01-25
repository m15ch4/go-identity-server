package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type TokenService interface {
	GenerateToken(user User) (string, error)
	ValidateToken(token string) (User, error)
	RejectToken(token string) error
	GetRejectedTokens() ([]string, error)
}

type tokenService struct {
	secret         string
	rejectedTokens []string
}

func NewTokenService(secret string) TokenService {
	return &tokenService{secret: secret}
}

// GenerateToken generates a JWT token for the given user.
// The token is valid for 5 minutes.
func (t *tokenService) GenerateToken(user User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["name"] = user.Name
	claims["exp"] = time.Now().Add(time.Minute * 5).Unix()
	// use "github.com/google/uuid" to generate a UUID and assign it to the "jti" claim
	claims["jti"] = uuid.New().String()
	tokenString, err := token.SignedString([]byte(t.secret))
	return tokenString, err
}

// ValidateToken validates the given JWT token.
// If the token is valid, it returns the user associated with the token.
// If the token is not valid, it returns an error.
func (t *tokenService) ValidateToken(tokenString string) (User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secret), nil
	})
	if err != nil {
		return User{}, err
	}
	// check if tokens jti claim is on rejected list
	for _, rejectedJTI := range t.rejectedTokens {
		if rejectedJTI == token.Claims.(jwt.MapClaims)["jti"].(string) {
			return User{}, fmt.Errorf("token is rejected")
		}
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return User{
				ID:   int(claims["user_id"].(float64)),
				Name: string(claims["name"].(string))},
			nil
	}
	return User{}, fmt.Errorf("invalid token")
}

// RejectToken decodes token and rejects the given JWT token by adding it's id to the list of rejected tokens.
func (t *tokenService) RejectToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secret), nil
	})
	if err != nil {
		return err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		t.rejectedTokens = append(t.rejectedTokens, string(claims["jti"].(string)))
	}
	return nil
}

// GetRejectedTokens returns the list of rejected tokens.
func (t *tokenService) GetRejectedTokens() ([]string, error) {
	return t.rejectedTokens, nil
}

type User struct {
	ID        int
	Name      string
	Password  string
	FirstName string
	LastName  string
	Role      string
	Age       int
}

type LoginUser struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Role      string `json:"role"`
	Age       int    `json:"age"`
}

type Server struct {
	tokenService TokenService
	users        []User
	port         int
	host         string
}

// NewServer creates a new server with the given token service, users and port.
// The host defaults to "localhost".
// The port defaults to 8080.
// The token service must not be nil.
func NewServer(tokenService TokenService, users []User, port int, host string) *Server {
	return &Server{tokenService: tokenService, users: users, port: port, host: host}
}

// Serve configures http routing using gin library and starts the server.
func (s *Server) Serve() {
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.POST("/login", s.login)
	r.GET("/", s.home)
	r.POST("/users", s.createUser)
	r.GET("/users", s.listUsers)
	r.POST("/reject", s.reject)
	r.GET("/rejected", s.listRejected)
	r.Run(fmt.Sprintf(":%d", s.port))
	fmt.Printf("Server listening on port %d\n", s.port)
}

func (s *Server) home(ctx *gin.Context) {
	// get authorization token from header
	tokenString := ctx.Request.Header.Get("Authorization")
	if tokenString == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "No token provided",
		})
		return
	}
	// strip "Bearer " from token string
	tokenString = tokenString[7:]

	// validate token
	user, err := s.tokenService.ValidateToken(tokenString)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}

	// get user info
	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Hello %s", user.Name),
	})
}

// createUser creates a new user with the given name and password.
func (s *Server) createUser(ctx *gin.Context) {
	// get authorization token from header
	tokenString := ctx.Request.Header.Get("Authorization")
	if tokenString == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "No token provided",
		})
		return
	}
	// strip "Bearer " from token string
	tokenString = tokenString[7:]

	// validate token
	_, err := s.tokenService.ValidateToken(tokenString)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}

	var createUserRequest CreateUserRequest
	if err := ctx.ShouldBindJSON(&createUserRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	newUser := User{
		ID:        len(s.users) + 1,
		Name:      createUserRequest.Name,
		Password:  createUserRequest.Password,
		FirstName: createUserRequest.FirstName,
		LastName:  createUserRequest.LastName,
		Role:      createUserRequest.Role,
		Age:       createUserRequest.Age,
	}
	s.users = append(s.users, newUser)
	// return user info
	ctx.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("User %s created", newUser.Name),
	})
}

// listUsers returns a list of users.
func (s *Server) listUsers(ctx *gin.Context) {
	// get authorization token from header
	tokenString := ctx.Request.Header.Get("Authorization")
	if tokenString == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "No token provided",
		})
		return
	}
	// strip "Bearer " from token string
	tokenString = tokenString[7:]

	// validate token
	_, err := s.tokenService.ValidateToken(tokenString)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"users": s.users,
	})
}

// login authenticates the given user and returns a JWT token.
func (s *Server) login(ctx *gin.Context) {
	// validate loginuser password and generate token
	var loginUser LoginUser
	if err := ctx.ShouldBindJSON(&loginUser); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	for _, user := range s.users {
		if loginUser.Name == user.Name && loginUser.Password == user.Password {
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
			return
		}
	}
	ctx.JSON(http.StatusUnauthorized, gin.H{
		"message": "Unauthorized",
	})
}

// reject a JWT token provided in the header.
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

// listRejected handler returns the list of rejected tokens.
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

// main initializes the services and starts the server on port 8080.
// it uses Gin to handle http routing and to serve the HTTP requests.
func main() {
	tokenService := NewTokenService("secret")
	users := []User{
		{ID: 1, Name: "John", Password: "VMware1!"},
		{ID: 2, Name: "Jane", Password: "VMware1!"},
	}
	server := NewServer(tokenService, users, 8080, "localhost")
	server.Serve()
}
