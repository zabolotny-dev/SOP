package main

import (
	"fmt"
	"hosting-contracts/api"
	"hosting-service/internal/domain"
	"hosting-service/internal/handlers"
	"hosting-service/internal/repository/psql"
	"hosting-service/internal/service"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {

	psqlDb, err := gorm.Open(postgres.Open("postgres://postgres:vladick@localhost:5432/sop?search_path=public"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Connection to DB failed : %v", err)
	}

	migrateTables(psqlDb)

	planRepository := psql.NewPlanRepository(psqlDb)
	serverRepository := psql.NewServerRepository(psqlDb)

	serverService := service.NewServerService(serverRepository)
	planService := service.NewPlanService(planRepository)

	apiHandler := handlers.NewApiHandler(planService, serverService)

	strictHandler := api.NewStrictHandler(apiHandler, nil)

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/swagger/doc.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Write(api.OpenApiSpec)
	})

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.yaml"),
	))

	handler := api.HandlerFromMux(strictHandler, router)

	port := "8080"
	fmt.Printf("Сервер успешно запущен. Адрес: http://localhost:%s\n", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}

func migrateTables(db *gorm.DB) {
	db.AutoMigrate(domain.Plan{})
	db.AutoMigrate(domain.Server{})
	//db.AutoMigrate(domain.User{})
}
