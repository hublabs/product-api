module github.com/hublabs/product-api

go 1.12

require (
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.1.0
	github.com/Shopify/sarama v1.26.1
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-xorm/xorm v0.7.9
	github.com/hublabs/common v0.0.0-20200322235728-74e36c68d78c
	github.com/labstack/echo v3.3.10+incompatible
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/pangpanglabs/echoswagger v1.2.0
	github.com/pangpanglabs/goutils v0.0.0-20200320140103-932a39405894
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli/v2 v2.2.0
)

replace github.com/go-xorm/xorm => github.com/pangpanglabs/xorm v0.6.7-0.20191028024856-98149f1c9e95
