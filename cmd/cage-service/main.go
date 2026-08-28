package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/animalcage/internal/httpapi"
	"example.com/animalcage/internal/model"
	"example.com/animalcage/internal/store"
	"example.com/animalcage/internal/workflow"
)

type storeAdapter struct{ db *store.Store }

func (a storeAdapter) SaveRecord(v model.Record) error { return a.db.SaveRecord(v) }
func (a storeAdapter) GetRecord(id string) (model.Record, bool, error) {
	return a.db.GetRecord(id)
}
func (a storeAdapter) ListRecords() ([]model.Record, error) { return a.db.ListRecords() }
func (a storeAdapter) SaveAudit(v model.AuditEvent) error   { return a.db.SaveAuditEvent(v) }
func (a storeAdapter) ListAudit(id string) ([]model.AuditEvent, error) {
	return a.db.ListAuditEventsForRecord(id)
}
func (a storeAdapter) SaveWorkflow(v model.Workflow) error { return a.db.SaveWorkflow(v) }
func (a storeAdapter) ListWorkflows(id string) ([]model.Workflow, error) {
	return a.db.ListWorkflowsForRecord(id)
}
func (a storeAdapter) SaveAttachment(v model.Attachment) error { return a.db.SaveAttachment(v) }
func (a storeAdapter) ListAttachments(id string) ([]model.Attachment, error) {
	return a.db.ListAttachmentsForRecord(id)
}

func main() {
	dbPath := flag.String("db", "cage-service.db", "bbolt database path")
	address := flag.String("address", ":8080", "HTTP listen address")
	flag.Parse()
	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer db.Close()
	clock := workflow.FixedClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	service := workflow.NewService(storeAdapter{db: db}, clock)
	server := &http.Server{Addr: *address, Handler: httpapi.NewHandler(service), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("cage service listening on %s", *address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
