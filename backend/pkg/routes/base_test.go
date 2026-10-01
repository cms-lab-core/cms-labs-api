package routes

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	resty "github.com/go-resty/resty/v2"
	"github.com/h2non/gock"

	"github.com/cms-lab-core/cms-labs-api/shared/logs"

	json "github.com/goccy/go-json"

	"github.com/cms-lab-core/cms-labs-api/backend/app/addons/vault"
	"github.com/cms-lab-core/cms-labs-api/backend/app/di"
	"github.com/cms-lab-core/cms-labs-api/backend/app/models"
	"github.com/cms-lab-core/cms-labs-api/backend/pkg/configs"
	"github.com/google/uuid"

	"github.com/cms-lab-core/cms-labs-api/backend/pkg/middleware"
	"github.com/cms-lab-core/cms-labs-api/backend/platform/database"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type TestHTTP struct {
	App *fiber.App
	DB  *gorm.DB
}

type TestHttpRequest struct {
	Method        string
	Route         string
	Body          io.Reader
	Authorization string
	ContentType   string
}

func (f *TestHTTP) Request(
	r *TestHttpRequest,
) (int, string) {
	req := httptest.NewRequest(r.Method, r.Route, r.Body)
	if r.ContentType == "" {
		r.ContentType = "application/json"
	}
	req.Header.Set("Content-Type", r.ContentType)
	req.Header.Set("Authorization", r.Authorization)

	// Perform the request plain with the app.
	resp, _ := f.App.Test(req, -1)
	responseString := ""
	if b, err := io.ReadAll(resp.Body); err == nil {
		responseString = string(b)
	}
	return resp.StatusCode, responseString
}

type TestRpcRequest struct {
	Method        string
	Params        interface{}
	Authorization string
	ID            string
}

func (f *TestHTTP) Rpc(
	r *TestRpcRequest,
) (int, string) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return f.Request(
		&TestHttpRequest{
			Method: "POST",
			Route:  "/api/v1/rpc",
			Body: FiberRequestPayload(
				fiber.Map{
					"jsonrpc": "2.0",
					"method":  r.Method,
					"params":  r.Params,
					"id":      r.ID,
				},
			),
			Authorization: r.Authorization,
			ContentType:   "application/json",
		},
	)
}

func (f *TestHTTP) AuthorizationUser(userID uint, serverID *uint) string {
	roleAdmin := models.Role{}
	roleAdmin.Code = "admin"
	roleAdmin.Name = "Admin"
	f.DB.Where("code = ?", roleAdmin.Code).First(&roleAdmin)
	if roleAdmin.ID == 0 {
		f.DB.Create(&roleAdmin)
	}

	roleInstructor := models.Role{}
	roleInstructor.Code = "instructor"
	roleInstructor.Name = "Instructor"
	f.DB.Where("code = ?", roleInstructor.Code).First(&roleInstructor)
	if roleInstructor.ID == 0 {
		f.DB.Create(&roleInstructor)
	}

	if userID == 0 {
		entity := models.User{}
		entity.Email = uuid.New().String() + "@admin.com"
		f.DB.Create(&entity)

		roleUserAdmin := models.RoleRelation{}
		roleUserAdmin.RoleID = roleAdmin.ID
		roleUserAdmin.UserID = &entity.ID
		f.DB.Create(&roleUserAdmin)

		roleUserInstructor := models.RoleRelation{}
		roleUserInstructor.RoleID = roleInstructor.ID
		roleUserInstructor.UserID = &entity.ID
		f.DB.Create(&roleUserInstructor)

		userID = entity.ID
	}
	container, _ := di.NewDIContainer(&logs.ZeroLoggerConf{})
	uc := container.AuthTokenManager()
	token, _ := uc.NewJWTByUserId("", userID, serverID, nil)
	return "Bearer " + token.AccessToken
}

func (f *TestHTTP) AuthorizationServiceBasic() (string, string) {
	entity := models.Server{}
	entity.Type = models.ServerTypeOpenID
	entity.Name = uuid.New().String() + "_server"
	entity.Url = "https://localhost"
	entity.IsActive = true
	entity.Token = uuid.New().String()
	entity.ClientID = uuid.New().String()
	f.DB.Create(&entity)
	return "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", entity.ClientID, entity.Token)),
	), entity.ClientID
}

func (f *TestHTTP) Close() error {
	if f.DB != nil {
		sqlDB, err := f.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func NewTestHTTP() *TestHTTP {
	// Load .env.test file from the root folder.
	_ = godotenv.Load("../../.env")
	if err := godotenv.Load("../../.env.test"); err != nil {
		panic(err)
	}
	// Load .env local file form the additional
	configs.AppConfig.Reload()

	// Inject Resty
	di.NewRestyClient = func() *resty.Client {
		return resty.NewWithClient(
			&http.Client{Transport: gock.DefaultTransport},
		)
	}

	// Inject Vault mock
	di.NewVaultClient = func() vault.ClientInterface {
		return &vault.MockVaultClient{}
	}

	// Define a new Fiber app.
	app := fiber.New()
	middleware.FiberMiddleware(app)
	FiberRoutes(app)
	db, _ := database.MysqlConnection(logs.NewZeroLogger(&logs.ZeroLoggerConf{}))
	return &TestHTTP{
		App: app,
		DB:  db,
	}
}

func FiberJSON(data interface{}) string {
	jsonByte, _ := json.Marshal(data)
	return string(jsonByte)
}

func FiberRequestPayload(data interface{}) io.Reader {
	return strings.NewReader(fmt.Sprint(FiberJSON(data)))
}

func FiberRequestFormPayload(data map[string]string) io.Reader {
	// Создаем form-data запрос
	form := url.Values{}
	for key, value := range data {
		form.Add(key, value)
	}
	return strings.NewReader(form.Encode())
}
