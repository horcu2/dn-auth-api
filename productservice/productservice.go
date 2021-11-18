package productservice

import (
	//"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/lib/pq"
	"github.com/shadowshot-x/micro-product-go/productservice/store"
	"go.uber.org/zap"
	"net/http"
	"os"
	"strconv"
	"time"
)

var db *gorm.DB

//func GetSecret() string {
//	return os.Getenv("MYSQL_SECRET")
//}

func GetPgSecret() string {
	return os.Getenv("PG_SECRET")
}

// ProductController is the getproduct route handler
type ProductController struct {
	logger *zap.Logger
}

const (
	host     = "35.223.83.195"
	port     = 5432
	user     = "postgres"
	password = "Rj1GhiWQUxsH"
	dbname   = "postgres"
)

// NewProductController returns a frsh Upload controller
func NewProductController(logger *zap.Logger) *ProductController {
	return &ProductController{
		logger: logger,
	}
}

//func (ctrl *ProductController) InitMySqlGormConnection() {
//	// database configuration for mysql
//	// first we fetch the mysql secret string stored in environment variables
//	mySqlsecret := GetSecret()
//	// if secret is empty, we want to warn the user
//	if mySqlsecret == "" {
//		ctrl.logger.Warn("Unable to get mysql secret")
//		return
//	}
//	var err error
//	// lets open the conncection
//	db, err = gorm.Open("mysql", GetSecret())
//	if err != nil {
//		ctrl.logger.Warn("Connection Failed to Open", zap.Error(err))
//	} else {
//		ctrl.logger.Info("Connection Established")
//	}
//
//	//We have the database name in our Environment secret.
//	// Auto Migrate creates a table named products in that Database
//	db.AutoMigrate(&store.Product{})
//}

func (ctrl *ProductController) InitPgGormConnection() {
	// database configuration for mysql
	// first we fetch the mysql secret string stored in environment variables
	pgsecret := GetPgSecret()
	// if secret is empty, we want to warn the user
	if pgsecret == "" {
		ctrl.logger.Warn("Unable to get postgres secret")
		return
	}

	//dsn := GetPgSecret()
	//db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	// db, err := sql.Open("postgres", psqlInfo)
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	db, err := gorm.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}
	//defer db.Close()

	fmt.Println("Successfully connected!")
	db.Debug().AutoMigrate(&store.Product{})

	//var err error
	//// lets open the conncection
	//db, err = gorm.Open("mysql", GetSecret())
	//if err != nil {
	//	ctrl.logger.Warn("Connection Failed to Open", zap.Error(err))
	//} else {
	//	ctrl.logger.Info("Connection Established")
	//}
	//
	////We have the database name in our Environment secret.
	//// Auto Migrate creates a table named products in that Database
	//db.AutoMigrate(&store.Product{})
}

func (ctrl *ProductController) GetAllProductsHandler(rw http.ResponseWriter, r *http.Request) {
	// we know we will get a list of all products.
	AllProducts := []store.Product{}
	// here db.Find fetches all the existing Elements in Products and stores them in AllProducts
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := gorm.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}
	db.Find(&AllProducts)
	defer db.Close()
	// We can Send back all values to the ResponseWriter by jsonencoding the results
	json.NewEncoder(rw).Encode(AllProducts)
}

func handleNotInHeader(rw http.ResponseWriter, r *http.Request, param string) {
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write([]byte(fmt.Sprintf("%s Missing", param)))
}

func (ctrl *ProductController) AddProductHandler(rw http.ResponseWriter, r *http.Request) {
	//validate the request first
	if _, ok := r.Header["Productname"]; !ok {
		ctrl.logger.Warn("Name was not found in the header")
		handleNotInHeader(rw, r, "name")
		return
	}
	if _, ok := r.Header["Productvendor"]; !ok {
		ctrl.logger.Warn("Vendor was not found in the header")
		handleNotInHeader(rw, r, "Vendor")
		return
	}
	if _, ok := r.Header["Productinventory"]; !ok {
		ctrl.logger.Warn("Inventory was not found in the header")
		handleNotInHeader(rw, r, "Inventory")
		return
	}
	if _, ok := r.Header["Productdescription"]; !ok {
		ctrl.logger.Warn("Description was not found in the header")
		handleNotInHeader(rw, r, "Description")
		return
	}
	// We want to get the details of the Product first. So these have to be in the request
	inventory, err := strconv.Atoi(r.Header["Productinventory"][0])
	if err != nil {
		ctrl.logger.Error("Error converting string to integer in inventory", zap.Error(err))
	}
	newProduct := store.Product{
		Name:        r.Header["Productname"][0],
		VendorName:  r.Header["Productvendor"][0],
		Inventory:   inventory,
		Description: r.Header["Productdescription"][0],
		CreateAt:    time.Now(),
	}
	ctrl.logger.Info("1 Product was Added")
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := gorm.Open("postgres", psqlInfo)

	if err != nil {
		panic(err)
	}
	db.Omit("Id").Create(&newProduct)
	ctrl.logger.Info("Product was Added")
	defer db.Close()
}
