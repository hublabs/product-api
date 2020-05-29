module github.com/hublabs/product-api

go 1.12

require (
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.1.0
	github.com/Shopify/sarama v1.26.1
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-xorm/xorm v0.7.9
	github.com/hublabs/common v0.0.0-20200517114719-fafc696cb4c2
	github.com/labstack/echo v3.3.10+incompatible
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/pangpanglabs/echoswagger v1.2.0
	github.com/pangpanglabs/goutils v0.0.0-20200320140103-932a39405894
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli/v2 v2.2.0
	github.com/valyala/fasttemplate v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37 // indirect
	golang.org/x/net v0.0.0-20200528225125-3c3fba18258b // indirect
	golang.org/x/sys v0.0.0-20200523222454-059865788121 // indirect
)

replace github.com/go-xorm/xorm => github.com/pangpanglabs/xorm v0.6.7-0.20191028024856-98149f1c9e95
