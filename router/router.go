package router

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gookit/color"
	middlewareContract "github.com/nd-tools/capyvel/contracts/middlewares"
	routerContract "github.com/nd-tools/capyvel/contracts/router"
	"github.com/nd-tools/capyvel/foundation"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	DefaultGroupPath                   = "/api"
	ErrInvalidTimezone                 = "Invalid app.timezone configuration: %v"
	ErrPrefixRequired                  = "Prefix name is required %s"
	ErrGroupNameRequired               = "Group name is required"
	ErrIncorrectHTTPMethod             = "Invalid HTTP method: %s"
	ErrPortMisconfigured               = "HTTP port is misconfigured"
	ErrTLSConfigError                  = "Error in TLS configuration"
	ErrTLSCertPathNotFound             = "TLS certificate path not found"
	ErrTLSKeyPathNotFound              = "TLS key path not found"
	ErrMissingOrInvalidTimezone        = "app.timezone missing or invalid"
	ErrInvalidTimezoneConfig           = "Invalid timezone configuration: %v"
	ErrMissingOrInvalidAppEnv          = "App environment configuration is invalid or missing"
	ErrMissingOrInvalidDebugConfig     = "Debug configuration is invalid or missing"
	ErrMissingOrInvalidCORSMethods     = "CORS allowed methods configuration is invalid or missing"
	ErrMissingOrInvalidCORSOrigins     = "CORS allowed origins configuration is invalid or missing"
	ErrMissingOrInvalidCORSHeaders     = "CORS allowed headers configuration is invalid or missing"
	ErrMissingOrInvalidCORSCredentials = "CORS supports credentials configuration is invalid or missing"
)

// RouterManager is the global router manager instance.
var (
	RouterManager Router
)

// Router manages the Gin engine, default group, and middleware stack.
type Router struct {
	engine       *gin.Engine                     // The Gin engine instance
	defaultRoute *gin.RouterGroup                // Default API route group
	middlewares  []middlewareContract.Middleware // List of registered middlewares
}

// RouteOptions defines configuration options for registering routes.
type RouteOptions struct {
	BasePath                  string                          // Base path for the route group
	GroupName                 string                          // Name of the route group
	DontUseDefaultMiddlewares bool                            // Whether to skip default middlewares
	Middlewares               []middlewareContract.Middleware // Middlewares specific to this route
	Resource                  *routerContract.Resource        // Resource configuration for CRUD endpoints
}

// RouteOptionFunction defines a single function route configuration.
type RouteOptionFunction struct {
	PrefixName                string                          // Prefix for the route path
	DontUseDefaultMiddlewares bool                            // Whether to skip default middlewares
	HttpMethod                string                          // HTTP method (GET, POST, etc.)
	Function                  func(*gin.Context)              // Function handler for the route
	Middlewares               []middlewareContract.Middleware // Middlewares specific to this route
}

// Boot initializes the router, CORS, and app configuration.
func Boot() {
	// Set the timezone for the application
	if name, ok := foundation.App.Config.Get("app.timezone", "America/Mexico_City").(string); ok {
		location, err := time.LoadLocation(name)
		if err != nil {
			color.Redf(ErrInvalidTimezoneConfig, err)
			os.Exit(1)
		}
		time.Local = location
	} else {
		color.Redln(ErrMissingOrInvalidTimezone)
		os.Exit(1)
	}

	// Set the Gin mode based on the app environment
	if mode, ok := foundation.App.Config.Get("app.env", "dev").(string); ok && strings.EqualFold(mode, "release") {
		gin.SetMode(gin.ReleaseMode)
	} else if !ok {
		color.Redln(ErrMissingOrInvalidAppEnv)
		os.Exit(1)
	}

	// Create a new Gin engine
	router := gin.New()

	// Enable recovery middleware in debug mode
	if debug, ok := foundation.App.Config.Get("app.debug", true).(bool); ok && debug {
		router.Use(gin.Recovery())
	} else if !ok {
		color.Redln(ErrMissingOrInvalidDebugConfig)
		os.Exit(1)
	}

	// Configure CORS settings
	config := cors.DefaultConfig()
	if methods, ok := foundation.App.Config.Get("cors.allowed_methods", []string{"*"}).([]string); ok {
		config.AllowMethods = methods
	} else {
		color.Redln(ErrMissingOrInvalidCORSMethods)
		os.Exit(1)
	}

	if origins, ok := foundation.App.Config.Get("cors.allowed_origins", []string{"*"}).([]string); ok {
		config.AllowOrigins = origins
	} else {
		color.Redln(ErrMissingOrInvalidCORSOrigins)
		os.Exit(1)
	}

	if headers, ok := foundation.App.Config.Get("cors.allowed_headers", []string{"*"}).([]string); ok {
		config.AllowHeaders = headers
	} else {
		color.Redln(ErrMissingOrInvalidCORSHeaders)
		os.Exit(1)
	}

	if credentials, ok := foundation.App.Config.Get("cors.supports_credentials", false).(bool); ok {
		config.AllowCredentials = credentials
	} else {
		color.Redln(ErrMissingOrInvalidCORSCredentials)
		os.Exit(1)
	}

	router.Use(cors.New(config))

	// Add a logger middleware to log incoming requests
	router.Use(requestLoggerMiddleware())

	RouterManager.engine = router
	RouterManager.defaultRoute = RouterManager.engine.Group(DefaultGroupPath)
}

// requestLoggerMiddleware logs all incoming requests with method, path, and timestamp
func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process the request
		c.Next()

		// Log the request details
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		fmt.Printf("[%s] %s %s %d %s\n",
			start.Format("2006-01-02 15:04:05"),
			method,
			path,
			statusCode,
			duration,
		)
	}
}

// RegisterDefaultsMiddlewares registers a list of default middlewares.
func (router *Router) RegisterDefaultsMiddlewares(middlewares []middlewareContract.Middleware) {
	router.middlewares = append(router.middlewares, middlewares...)
}

// RegisterResource registers a set of CRUD routes for a resource controller.
func (router *Router) RegisterResource(option RouteOptions, controller routerContract.ResourceController) {
	r := RouterManager.defaultRoute
	if option.BasePath != "" {
		r = router.engine.Group(option.BasePath)
	}
	if option.GroupName == "" {
		color.Redln(ErrGroupNameRequired)
		os.Exit(1)
	}

	r = r.Group(option.GroupName)

	if !option.DontUseDefaultMiddlewares {
		for _, middleware := range router.middlewares {
			r.Use(middleware.Middleware)
		}
	}

	for _, middleware := range option.Middlewares {
		r.Use(middleware.Middleware)
	}

	if option.Resource == nil {
		r.GET("/", controller.Index)
		r.POST("/", controller.Store)
		r.GET("/:id", controller.Show)
		r.PUT("/:id", controller.Update)
		r.DELETE("/:id", controller.Destroy)
	} else {
		if option.Resource.Index {
			r.GET("/", controller.Index)
		}
		if option.Resource.Store {
			r.POST("/", controller.Store)
		}
		if option.Resource.Show {
			r.GET("/:id", controller.Show)
		}
		if option.Resource.Update {
			r.PUT("/:id", controller.Update)
		}
		if option.Resource.Destroy {
			r.DELETE("/:id", controller.Destroy)
		}
	}
}

// RegisterFunctions registers custom routes with specific handlers and HTTP methods.
func (router *Router) RegisterFunctions(option RouteOptions, optionsFunctions []RouteOptionFunction) {
	r := RouterManager.defaultRoute
	if option.BasePath != "" {
		r = router.engine.Group(option.BasePath)
	}
	if option.GroupName == "" {
		color.Redln(ErrGroupNameRequired)
		os.Exit(1)
	}
	r = r.Group(option.GroupName)

	for _, optionFunction := range optionsFunctions {
		httpMethod := optionFunction.HttpMethod
		function := optionFunction.Function
		prefixName := optionFunction.PrefixName

		if prefixName == "" {
			color.Redf(ErrPrefixRequired, getFunctionName(optionFunction.Function))
			os.Exit(1)
		}
		fullPath := "/" + prefixName
		middlewares := []gin.HandlerFunc{}
		if !option.DontUseDefaultMiddlewares && !optionFunction.DontUseDefaultMiddlewares {
			for _, middleware := range router.middlewares {
				middlewares = append(middlewares, middleware.Middleware)
			}
			for _, middleware := range option.Middlewares {
				middlewares = append(middlewares, middleware.Middleware)
			}
		}
		for _, middleware := range optionFunction.Middlewares {
			middlewares = append(middlewares, middleware.Middleware)
		}
		switch httpMethod {
		case http.MethodGet:
			r.GET(fullPath, append(middlewares, function)...)
		case http.MethodPost:
			r.POST(fullPath, append(middlewares, function)...)
		case http.MethodPut:
			r.PUT(fullPath, append(middlewares, function)...)
		case http.MethodDelete:
			r.DELETE(fullPath, append(middlewares, function)...)
		case http.MethodOptions:
			r.OPTIONS(fullPath, append(middlewares, function)...)
		default:
			color.Redf(ErrIncorrectHTTPMethod, httpMethod)
			os.Exit(1)
		}
	}
}

// Run starts the Gin server on the configured port with optional TLS.
func (router *Router) Run() *gin.Engine {
	config := foundation.App.Config
	port, ok := config.Get("http.port", 8080).(int)
	if !ok {
		color.Redln(ErrPortMisconfigured)
		os.Exit(1)
	}
	addr := fmt.Sprintf(":%d", port)
	runtls, ok := config.Get("http.tls.enable", false).(bool)
	if !ok {
		color.Redln(ErrTLSConfigError)
		os.Exit(1)
	}

	if runtls {
		certFile, ok := config.Get("http.tls.ssl.cert", "").(string)
		if !ok {
			color.Redln(ErrTLSCertPathNotFound)
			os.Exit(1)
		}

		keyFile, ok := config.Get("http.tls.ssl.key", "").(string)
		if !ok {
			color.Redln(ErrTLSKeyPathNotFound)
			os.Exit(1)
		}
		RouterManager.engine.RunTLS(addr, certFile, keyFile)
	} else {
		RouterManager.engine.Run(addr)
	}
	return RouterManager.engine
}

// getFunctionName returns the name of a given function.
func getFunctionName(function interface{}) string {
	ptr := reflect.ValueOf(function).Pointer()
	funcInfo := runtime.FuncForPC(ptr)
	if funcInfo != nil {
		return funcInfo.Name()
	}
	return "unknown"
}
