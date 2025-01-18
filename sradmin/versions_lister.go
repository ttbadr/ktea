package sradmin

import (
	tea "github.com/charmbracelet/bubbletea"
	"sync"
)

type Schema struct {
	Id      string
	Schema  string
	Version int
	Err     error
}

type SchemasListed struct {
	Schemas []Schema
}

type SchemaListingStarted struct {
	schemaChan   chan Schema
	versionCount int
}

func (s *SchemaListingStarted) AwaitCompletion() tea.Msg {
	var schemas []Schema
	count := 0

	for count < s.versionCount {
		select {
		case schema, ok := <-s.schemaChan:
			if !ok {
				// Channel is closed; exit loop
				return SchemasListed{Schemas: schemas}
			}
			schemas = append(schemas, schema)
			count++
		}
	}

	return SchemasListed{Schemas: schemas}
}

func (s *SrAdmin) ListVersions(subject string, versions []int) tea.Msg {
	schemaChan := make(chan Schema, len(versions))
	var wg sync.WaitGroup

	wg.Add(len(versions)) // Ensure this is outside the goroutine
	for _, version := range versions {
		go func(version int) {
			defer wg.Done() // Ensure Done() is called even on panic or early return
			schema, err := s.client.GetSchemaByVersion(subject, version)
			if err == nil {
				schemaChan <- Schema{
					Id:      schema.Schema(),
					Schema:  schema.Schema(),
					Version: version,
				}
			} else {
				schemaChan <- Schema{
					Err:     err,
					Version: version,
				}
			}
		}(version)
	}

	go func() {
		wg.Wait()         // Wait for all goroutines to finish
		close(schemaChan) // Close the channel to signal completion
	}()

	return SchemaListingStarted{
		schemaChan: schemaChan, versionCount: len(versions),
	}
}
