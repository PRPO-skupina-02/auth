package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/PRPO-skupina-02/auth/api"
	"github.com/PRPO-skupina-02/auth/db"
	"github.com/PRPO-skupina-02/common/database"
	"github.com/PRPO-skupina-02/common/logging"
	"github.com/PRPO-skupina-02/common/validation"
	"github.com/gin-gonic/gin"
)

func main() {
	err := run()

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func run() error {
	slog.Info("Starting server")

	logger := logging.GetDefaultLogger()
	slog.SetDefault(logger)

	db, err := database.OpenAndMigrateProd(db.MigrationsFS)
	if err != nil {
		return err
	}

	trans, err := validation.RegisterValidation()
	if err != nil {
		return err
	}

	router := gin.Default()
	api.Register(router, db, trans)

	slog.Info("Server startup complete")
	err = router.Run(":8080")
	if err != nil {
		return err
	}

	return nil
}
