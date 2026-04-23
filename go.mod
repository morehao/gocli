module github.com/morehao/gocli

go 1.26.1

require (
	github.com/morehao/golib v0.1.0
	github.com/spf13/cobra v1.10.2
	golang.org/x/mod v0.34.0
	gorm.io/driver/mysql v1.6.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1
)

replace github.com/morehao/golib => ../golib

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.9.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
