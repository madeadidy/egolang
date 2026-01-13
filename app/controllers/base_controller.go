package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"

	"github.com/codeuiprogramming/e-commerce/app/models"
	"github.com/codeuiprogramming/e-commerce/database/seeders"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/urfave/cli"
	"gorm.io/gorm"
)

type Server struct {
	DB *gorm.DB
	Router *mux.Router
	AppConfig *AppConfig
}

type AppConfig struct {
	AppName string
	AppEnv string
	AppPort string
	AppURL string
}

type DBConfig struct {
	DBHost string
	DBUser string
	DBPassword string
	DBName string
	DBPort string
}

type PageLink struct {
	Page int32
	Url string
	IsCurrentPage bool
}

type PaginationLinks struct {
	CurrentPage string
	NextPage string
	PrevPage string
	TotalRows int32
	TotalPages int32
	Links []PageLink
}

type PaginationParams struct {
	Path string
	TotalRows int32
	PerPage int32
	CurrentPage int32
}

type Result struct {
	Code int `json:"code"`
	Data interface{} `json:"data"`
	Message string `json:"message"`
}

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var sessionShoppingCart = "shopping-cart-session"
var sessionFlash = "flash-session"
var sessionUser = "user-session"

func (server *Server) Initialize(appConfig AppConfig, dbConfig DBConfig) {

	fmt.Println("Welcome to " + appConfig.AppName)

	server.initializeDB(dbConfig)
	server.initializeAppConfig(appConfig)
	server.initializeRoutes()
} 

func (server *Server) Run (addr string) {
	fmt.Printf("Listening to port %s", addr)
	log.Fatal(http.ListenAndServe(addr, server.Router))
}

func (server *Server) initializeDB(dbConfig DBConfig) {
	var err error
	// Membuat DSN (Data Source Name) berisi informasi untuk koneksi ke database PostgreSQL.
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", 
	dbConfig.DBHost, dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBName, dbConfig.DBPort)
	// Membuka koneksi ke database menggunakan GORM dengan driver PostgreSQL.
	server.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	 // Jika terjadi error saat menghubungkan database, hentikan server dan tampilkan panic.
	if err != nil {
		panic("Failed on connecting to the database server")
	}
}

func (server *Server) initializeAppConfig(appConfig AppConfig) {
	server.AppConfig = &appConfig
}

func (server *Server) dbMigrate() {
	for _, model := range models.RegisterModels() {
		err := server.DB.Debug().AutoMigrate(model.Model)

		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Database migrate successfully")
}

func (server *Server) InitCommands(config AppConfig, dbConfig DBConfig) {
	server.initializeDB(dbConfig)

	cmdApp := cli.NewApp()
	cmdApp.Commands = []cli.Command{
		{
			Name: "db:migrate",
			Action: func(c *cli.Context) error {
				server.dbMigrate()
				return nil
			},
		},
		{
			Name: "db:seed",
			Action: func(c *cli.Context) error {
				err := seeders.DBSeed(server.DB)
				if err != nil {
					log.Fatal(err)
				}

				return nil
			},
		},
	}

	err := cmdApp.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}


func GetPaginationLinks(config *AppConfig, params PaginationParams) (PaginationLinks, error) {
	var links []PageLink

	totalPages := int32(math.Ceil(float64(params.TotalRows) / float64(params.PerPage)))
	
	for i :=1; int32(i) <= totalPages; i++ {
		links = append(links, PageLink{
			Page: int32(i),
			Url: fmt.Sprintf("%s/%s?page=%s", config.AppURL, params.Path, fmt.Sprint(i)),
			IsCurrentPage: int32(i) == params.CurrentPage,
		})
	}

	var nextPage int32
	var prevPage int32

	prevPage = 1
	nextPage = totalPages

	if params.CurrentPage > 2 {
		prevPage = params.CurrentPage - 1
	}

	if params.CurrentPage < totalPages {
		nextPage = params.CurrentPage + 1
	}

	return PaginationLinks{
		CurrentPage: fmt.Sprintf("%s/%s?page=%s", config.AppURL, params.Path, fmt.Sprint(params.CurrentPage)),
		NextPage:    fmt.Sprintf("%s/%s?page=%s", config.AppURL, params.Path, fmt.Sprint(nextPage)),
		PrevPage:    fmt.Sprintf("%s/%s?page=%s", config.AppURL, params.Path, fmt.Sprint(prevPage)),
		TotalRows:   params.TotalRows,
		TotalPages:  totalPages,
		Links:       links,
	}, nil
}

func (server *Server) GetProvince() ([]models.Province, error) {
	req, err := http.NewRequest("GET", os.Getenv("API_ONGKIR_BASE_URL")+"destination/province", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Key", os.Getenv("API_ONGKIR_KEY"))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	var result models.ProvinceResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (server *Server) GetCitiesByProvinceID(provinceID string) ([]models.City, error) {
    url := os.Getenv("API_ONGKIR_BASE_URL") + "destination/city/" + provinceID

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Add("Key", os.Getenv("API_ONGKIR_KEY"))

    client := &http.Client{}
    res, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    body, _ := io.ReadAll(res.Body)
    fmt.Println("RAW RESPONSE:", string(body)) // debug body

    var result models.CityResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, err
    }

    return result.Data, nil
}

func (server *Server) CalculateShippingFee(shippingParams models.ShippingFeeParams) ([]models.ShippingFeeOption, error) {
	if shippingParams.Origin == "" || shippingParams.Destination == "" || shippingParams.Weight <= 0 || shippingParams.Courier == "" {
		return nil, errors.New("invalid params")
	}

	params := url.Values{}
	params.Add("origin", strings.TrimSpace(shippingParams.Origin))
	params.Add("destination", strings.TrimSpace(shippingParams.Destination))
	params.Add("weight", strconv.Itoa(shippingParams.Weight))
	params.Add("courier", strings.TrimSpace(shippingParams.Courier))

	endpoint := os.Getenv("API_ONGKIR_BASE_URL") + "calculate/domestic-cost"
	API_KEY := os.Getenv("API_ONGKIR_KEY")

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Key", API_KEY)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	log.Println("🚀 Raw Ongkir Response:", string(body))

	var ongkirResponse models.OngkirResponse
	if err := json.Unmarshal(body, &ongkirResponse); err != nil {
		return nil, err
	}

	if ongkirResponse.Meta.Code != 200 {
		return nil, fmt.Errorf("API error: %s", ongkirResponse.Meta.Message)
	}

	var shippingFeeOptions []models.ShippingFeeOption
	for _, data := range ongkirResponse.Data {
		shippingFeeOptions = append(shippingFeeOptions, models.ShippingFeeOption{
			Service: data.Service + " (" + data.Description + ")",
			Fee:     data.Cost,
		})
	}

	return shippingFeeOptions, nil
}

func SetFlash(w http.ResponseWriter, r *http.Request, name string, value string) {
	session, err := store.Get(r, sessionFlash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.AddFlash(value, name)
	session.Save(r, w)
}

func GetFlash(w http.ResponseWriter, r *http.Request, name string) []string {
	session, err := store.Get(r, sessionFlash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	fm := session.Flashes(name)
	if len(fm) == 0 {
		return nil
	}

	session.Save(r, w)
	var flashes []string
	for _, fl := range fm {
		flashes = append(flashes, fl.(string))
	}

	return flashes
}

func IsLoggedIn(r *http.Request) bool {
	session, _ := store.Get(r, sessionUser)
	return session.Values["id"] != nil
}

func ComparePassword(password string, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) ==  nil
}

func MakePassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return string(hashedPassword), err
}

func (server *Server) CurrentUser(w http.ResponseWriter, r *http.Request) *models.User {
	if !IsLoggedIn(r) {
		return nil
	}

	session, _ := store.Get(r, sessionUser)

	userModel := models.User{}
	user, err := userModel.FindByID(server.DB, session.Values["id"].(string))
	if err != nil {
		session.Values["id"] = nil
		session.Save(r, w)
		return nil
	}

	return user
}

// DefaultRenderData menggabungkan data template umum seperti Clerk publishable key dan user saat ini.
func (server *Server) DefaultRenderData(w http.ResponseWriter, r *http.Request, data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = map[string]interface{}{}
	}

	// Tambahkan Clerk publishable key dari environment
	data["ClerkPublishableKey"] = os.Getenv("CLERK_PUBLISHABLE_KEY")

	// Tambahkan user saat ini (masih menggunakan session lama jika ada)
	data["user"] = server.CurrentUser(w, r)

	return data
}