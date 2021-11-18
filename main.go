package main

import (
	"github.com/shadowshot-x/micro-product-go/authservice"
	"github.com/shadowshot-x/micro-product-go/authservice/middleware"
	"github.com/shadowshot-x/micro-product-go/clientclaims"
	"github.com/shadowshot-x/micro-product-go/productservice"
	"html/template"
	"log"
	"net/http"

	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type templateData struct {
	Message string
}

var (
	data templateData
	tmpl *template.Template
)

func main() {
	log, _ := zap.NewProduction()
	defer func(log *zap.Logger) {
		err := log.Sync()
		if err != nil {

		}
	}(log)

	log.Info("Starting...")

	err := godotenv.Load(".env")

	if err != nil {
		log.Error("Error loading .env file", zap.Error(err))
	}

	mainRouter := mux.NewRouter()

	suc := authservice.NewSignupController(log)
	sic := authservice.NewSigninController(log)
	uc := clientclaims.NewUploadController(log)
	dc := clientclaims.NewDownloadController(log)
	tm := middleware.NewTokenMiddleware(log)
	pc := productservice.NewProductController(log)

	// readiness
	homeRouter := mainRouter.PathPrefix("/").Subrouter()
	homeRouter.HandleFunc("/health", home).Methods("GET")

	// We will create a Subrouter for Authentication service
	// route for sign up and signin. The Function will come from auth-service package
	// checks if given header params already exists. If not,it adds the user
	authRouter := mainRouter.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/signup", suc.SignupHandler).Methods("POST")

	// The Signin will send the JWT Token back as we are making microservices.
	// JWT token will make sure that other services are protected.
	// So, ultimately, we would need a middleware
	authRouter.HandleFunc("/signin", sic.SigninHandler).Methods("GET")

	// File Upload SubRouter
	claimsRouter := mainRouter.PathPrefix("/claims").Subrouter()
	claimsRouter.HandleFunc("/upload", uc.UploadFile)
	claimsRouter.HandleFunc("/download", dc.DownloadFile)
	claimsRouter.Use(tm.TokenValidationMiddleware)

	//Initialize the Gorm connection
	pc.InitPgGormConnection()
	productRouter := mainRouter.PathPrefix("/product").Subrouter()
	productRouter.HandleFunc("/getprods", pc.GetAllProductsHandler).Methods("GET")
	productRouter.HandleFunc("/addprod", pc.AddProductHandler).Methods("POST")

	// CORS Header
	ch := gohandlers.CORS(gohandlers.AllowedOrigins([]string{"http://localhost:3000"}))
	// Add the Middleware to different subrouter
	// HTTP Server
	// Add Time outs
	server := &http.Server{
		Addr:    "127.0.0.1:9092",
		Handler: ch(mainRouter),
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Error("Error Booting the Server", zap.Error(err))
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	log.Printf("Hello from Cloud Code! Received request: %s %s", r.Method, r.URL.Path)
	if err := tmpl.Execute(w, data); err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		log.Printf("template.Execute: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}
