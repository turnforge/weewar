package gormbe

import (
	"errors"
	"log"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const name = "github.com/panyam/onehub"

var (
	Tracer = otel.Tracer(name)
	Meter  = otel.Meter(name)
	Logger = otelslog.NewLogger(name)
)

var InvalidIDError = errors.New("ID is invalid or empty")
var MessageUpdateFailed = errors.New("Update failed concurrency check")
var TopicUpdateFailed = errors.New("Update failed concurrency check")
var UserUpdateFailed = errors.New("Update failed concurrency check")

func OpenDB(db_endpoint string) (db *gorm.DB, err error) {
	log.Println("Connecting to DB: ", db_endpoint)
	if strings.HasPrefix(db_endpoint, "postgres://") {
		db, err = gorm.Open(postgres.Open(db_endpoint), &gorm.Config{})
		if err != nil {
			panic(err)
		}
		/*
			} else if strings.HasPrefix(db_endpoint, "sqlite://") {
				dbpath := utils.ExpandUserPath((db_endpoint)[len("sqlite://"):])
				db, err = gorm.Open(sqlite.Open(dbpath), &gorm.Config{})
		*/
	}
	if err != nil {
		log.Println("Cannot connect DB: ", db_endpoint, err)
	} else {
		log.Println("Successfully connected DB: ", db_endpoint)
	}
	return
}

func OpenLilBattleDB(dbEndpoint string, defaultEndpoint string) *gorm.DB {
	if dbEndpoint == "" {
		dbEndpoint = os.Getenv("LILBATTLE_DB_ENDPOINT")
		if dbEndpoint == "" {
			dbEndpoint = defaultEndpoint
		}
	}
	db, err := OpenDB(dbEndpoint)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return db
}
