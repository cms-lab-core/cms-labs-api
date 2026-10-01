package routes

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/cms-lab-core/cms-labs-api/shared/connection"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/app/di"
	resty "github.com/go-resty/resty/v2"
	"github.com/h2non/gock"

	"github.com/cms-lab-core/cms-labs-api/shared/logs"

	json "github.com/goccy/go-json"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/pkg/configs"

	"github.com/cms-lab-core/cms-labs-api/pnetlabaddon/pkg/middleware"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type FiberTestHTTP struct {
	App *fiber.App
	DB  *gorm.DB
}

type FiberTestHttpRequest struct {
	Method        string
	Route         string
	Body          io.Reader
	Authorization string
	ContentType   string
}

func (f *FiberTestHTTP) Request(
	r *FiberTestHttpRequest,
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

func NewFiberTestHTTP() *FiberTestHTTP {
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

	// Define a new Fiber app.
	app := fiber.New()
	middleware.FiberMiddleware(app)
	FiberRoutes(app)
	db, _ := connection.MysqlConnection(configs.AppConfig.DB, logs.NewZeroLogger(&logs.ZeroLoggerConf{}))
	return &FiberTestHTTP{
		App: app,
		DB:  db,
	}
}

func FiberJSON(data any) string {
	jsonByte, _ := json.Marshal(data)
	return string(jsonByte)
}

func FiberRequestPayload(data any) io.Reader {
	return strings.NewReader(fmt.Sprint(FiberJSON(data)))
}
