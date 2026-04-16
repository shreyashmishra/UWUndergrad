package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"planahead/planner-api/internal/config"
	"planahead/planner-api/internal/db"
	"planahead/planner-api/internal/graph"
	"planahead/planner-api/internal/repository"
	"planahead/planner-api/internal/service"
)

func main() {
	cfg := config.Load()

	sqlDB, err := db.Open(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer sqlDB.Close()

	if err := db.BootstrapMySQL(context.Background(), sqlDB, cfg.MockUserID); err != nil {
		log.Fatalf("bootstrap mysql: %v", err)
	}

	catalogRepository := repository.NewCatalogRepository(sqlDB)
	studentRepository := repository.NewStudentRepository(sqlDB)
	programService := service.NewProgramService(catalogRepository, studentRepository, cfg.MockUserID)
	studentService := service.NewStudentService(studentRepository, cfg.MockUserID)
	resolver := graph.NewRootResolver(programService, studentService)

	schema := graphql.MustParseSchema(graph.Schema, resolver)

	router := chi.NewRouter()
	router.Use(graph.CORSMiddleware(cfg.AllowedOrigin))
	router.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("PlanAhead Go API"))
	})
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	router.Handle("/graphql", &relay.Handler{Schema: schema})

	addr := ":" + cfg.Port
	log.Printf("planner api listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil && err != http.ErrServerClosed {
		log.Printf("server failed: %v", err)
		os.Exit(1)
	}
}
