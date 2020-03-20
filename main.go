package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/hublabs/common/api"
	"github.com/hublabs/common/auth"
	"github.com/hublabs/product-api/adapters"
	"github.com/hublabs/product-api/config"
	"github.com/hublabs/product-api/controllers"
	"github.com/hublabs/product-api/factory"
	"github.com/hublabs/product-api/models"

	"github.com/asaskevich/govalidator"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pangpanglabs/echoswagger"
	"github.com/pangpanglabs/goutils/behaviorlog"
	"github.com/pangpanglabs/goutils/echomiddleware"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	c := config.Init(os.Getenv("APP_ENV"))
	db, err := initDB(c.Database.Driver, c.Database.Connection, c.Debug)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	factory.InitDB(db)

	if err := adapters.SetupMessagePublisher(c.EventBroker.Kafka); err != nil {
		panic(err)
	}
	defer adapters.MessagePublisher{}.Close()

	app := cli.NewApp()
	app.Name = "Hublabs Product Application"
	app.Commands = []*cli.Command{
		{
			Name:  "api-server",
			Usage: "run api server",
			Action: func(cliContext *cli.Context) error {
				e := echo.New()
				r := echoswagger.New(e, "doc", &echoswagger.Info{
					Title:       "Product API",
					Description: "This is docs for product-api service",
					Version:     "1.0.0",
				})

				r.AddSecurityAPIKey("Authorization", "JWT token", echoswagger.SecurityInHeader)
				r.SetUI(echoswagger.UISetting{
					HideTop: true,
				})

				e.GET("/ping", func(c echo.Context) error {
					return c.String(http.StatusOK, "pong")
				})

				controllers.BrandController{}.Init(r.Group("Brands", "v1/brands"))
				controllers.ProductController{}.Init(r.Group("Products", "v1/products"))
				controllers.SkuController{}.Init(r.Group("Skus", "v1/skus"))
				controllers.PriceController{}.Init(r.Group("Prices", "v1/prices"))
				e.Pre(middleware.RemoveTrailingSlash())
				e.Pre(echomiddleware.ContextBase())
				e.Use(middleware.Recover())
				e.Use(middleware.CORS())
				e.Use(echomiddleware.BehaviorLogger(c.ServiceName, c.BehaviorLog.Kafka))
				e.Use(echomiddleware.ContextDB(c.ServiceName, db, c.Database.Logger.Kafka))
				e.Use(auth.UserClaimMiddleware("/ping", "/doc"))

				e.Validator = &Validator{}
				e.Debug = c.Debug

				if e.Debug {
					behaviorlog.SetLogLevel(logrus.InfoLevel)
				}

				api.SetErrorMessagePrefix(c.ServiceName)

				if err := e.Start(":" + c.HttpPort); err != nil {
					log.Println(err)
				}
				return nil
			},
		}, {
			Name:  "export",
			Usage: "export from 3rd part",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "target",
					Aliases: []string{"t"},
					Usage:   "export for brand,product,sku",
				},
			},
			Action: func(cliContext *cli.Context) error {
				// TODO: crawler from 3rd part interface...
				return nil
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func initDB(driver, connection string, debug bool) (*xorm.Engine, error) {
	db, err := xorm.NewEngine(driver, connection)
	if err != nil {
		return nil, err
	}

	if driver == "sqlite3" {
		runtime.GOMAXPROCS(1)
	}

	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Minute * 10)

	db.ShowSQL(debug)
	if err := models.Init(db); err != nil {
		return nil, err
	}
	return db, nil
}

type Validator struct{}

func (v *Validator) Validate(i interface{}) error {
	_, err := govalidator.ValidateStruct(i)
	return err
}
