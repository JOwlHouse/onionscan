package crawldb

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/JOwlHouse/onionscan/pkg/model"
	"gorm.io/gorm"
)

// CrawlDB is the main interface for persistent storage in OnionScan
type CrawlDB struct {
	db *gorm.DB
}

// Initialize sets up a new database - should only be called when creating a
// new database.
func LoadDB() {
	log.Printf("Creating Craw...")

	log.Printf("Database Setup Complete")
}

// InsertCrawlRecord adds a new spider entry to the database and returns the
// record id.
func (cdb *CrawlDB) InsertCrawlRecord(url string, page *model.Page) (uint, error) {
	crawl := CrawlRecord{
		URL:       url,
		Timestamp: time.Now(),
		Page:      *page,
	}
	cdb.db.Create(&crawl)
	return crawl.ID, nil
}

// GetCrawlRecord returns a CrawlRecord from the database given an ID.
func (cdb *CrawlDB) GetCrawlRecord(id int) (CrawlRecord, error) {
	crawls := cdb.db.Use("crawls")
	readBack, err := crawls.Read(id)
	if err == nil {
		out, err := json.Marshal(readBack)
		if err == nil {
			var crawlRecord CrawlRecord
			json.Unmarshal(out, &crawlRecord)
			return crawlRecord, nil
		}
		return CrawlRecord{}, err
	}
	return CrawlRecord{}, err
}

// HasCrawlRecord returns true if a given URL is associated with a crawl record
// in the database. Only records created after the given duration are considered.
func (cdb *CrawlDB) HasCrawlRecord(url string, duration time.Duration) (bool, int) {
	var query interface{}
	before := time.Now().Add(duration)

	q := fmt.Sprintf(`{"eq":"%v", "in": ["URL"]}`, url)
	json.Unmarshal([]byte(q), &query)

	queryResult := make(map[int]struct{}) // query result (document IDs) goes into map keys
	crawls := cdb.db.Use("crawls")
	if err := db.EvalQuery(query, crawls, &queryResult); err != nil {
		panic(err)
	}

	for id := range queryResult {
		// To get query result document, simply read it
		readBack, err := crawls.Read(id)
		if err == nil {
			out, err := json.Marshal(readBack)
			if err == nil {
				var crawlRecord CrawlRecord
				json.Unmarshal(out, &crawlRecord)

				if crawlRecord.Timestamp.After(before) {
					return true, id
				}
			}
		}

	}

	return false, 0
}

// InsertRelationship creates a new Relationship in the database.
func (cdb *CrawlDB) InsertRelationship(onion string, from string, identiferType string, identifier string) (int, error) {

	rels, err := cdb.GetRelationshipsWithOnion(onion)

	// If we have seen this before, we will update rather than adding a
	// new relationship
	if err == nil {
		for _, rel := range rels {
			if rel.From == from && rel.Identifier == identifier && rel.Type == identiferType {
				// Update the Relationships
				log.Printf("Updating %s --- %s ---> %s (%s)", onion, from, identifier, identiferType)
				relationships := cdb.db.Use("relationships")
				err := relationships.Update(rel.ID, map[string]interface{}{
					"Onion":      onion,
					"From":       from,
					"Type":       identiferType,
					"Identifier": identifier,
					"FirstSeen":  rel.FirstSeen,
					"LastSeen":   time.Now()})
				return rel.ID, err
			}
		}
	}

	// Otherwise Insert New
	log.Printf("Inserting %s --- %s ---> %s (%s)", onion, from, identifier, identiferType)
	relationships := cdb.db.Use("relationships")
	docID, err := relationships.Insert(map[string]interface{}{
		"Onion":      onion,
		"From":       from,
		"Type":       identiferType,
		"Identifier": identifier,
		"FirstSeen":  time.Now(),
		"LastSeen":   time.Now()})
	return docID, err
}

// GetRelationshipsWithOnion returns all relationships with an Onion field matching
// the onion parameter.
func (cdb *CrawlDB) GetRelationshipsWithOnion(onion string) ([]Relationship, error) {
	return cdb.queryDB("Onion", onion)
}

// GetUserRelationshipFromOnion reconstructs a user relationship from a given
// identifier. fromonion is used as a filter to ensure that only user relationships
// from a given onion are reconstructed.
func (cdb *CrawlDB) GetUserRelationshipFromOnion(identifier string, fromonion string) (map[string]Relationship, error) {
	results, err := cdb.GetRelationshipsWithOnion(identifier)

	if err != nil {
		return nil, err
	}

	relationships := make(map[string]Relationship)
	for _, result := range results {
		if result.From == fromonion {
			relationships[result.Type] = result
		}
	}
	return relationships, nil
}

// GetAllRelationshipsCount returns the total number of relationships stored in
// the database.
func (cdb *CrawlDB) GetAllRelationshipsCount() int {
	queryResult := make(map[int]struct{}) // query result (document IDs) goes into map keys
	relationships := cdb.db.Use("relationships")

	if err := db.EvalAllIDs(relationships, &queryResult); err != nil {
		return 0
	}
	return len(queryResult)
}

// GetRelationshipsCount returns the total number of relationships for a given
// identifier.
func (cdb *CrawlDB) GetRelationshipsCount(identifier string) int {
	var query interface{}

	q := fmt.Sprintf(`{"eq":"%v", "in": ["Identifier"]}`, identifier)
	json.Unmarshal([]byte(q), &query)

	queryResult := make(map[int]struct{}) // query result (document IDs) goes into map keys
	relationships := cdb.db.Use("relationships")
	if err := db.EvalQuery(query, relationships, &queryResult); err != nil {
		return 0
	}
	return len(queryResult)
}

// GetRelationshipsWithIdentifier returns all relatioships associated with a
// given identifier.
func (cdb *CrawlDB) GetRelationshipsWithIdentifier(identifier string) ([]Relationship, error) {

	types, _ := cdb.queryDB("Type", identifier)
	froms, _ := cdb.queryDB("From", identifier)
	identifiers, _ := cdb.queryDB("Identifier", identifier)

	queryResult := append(types, froms...)
	queryResult = append(queryResult, identifiers...)

	return queryResult, nil
}

func (cdb *CrawlDB) queryDB(field string, value string) ([]Relationship, error) {
	var query interface{}

	q := fmt.Sprintf(`{"eq":"%v", "in": ["%v"]}`, value, field)
	json.Unmarshal([]byte(q), &query)

	queryResult := make(map[int]struct{}) // query result (document IDs) goes into map keys
	relationships := cdb.db.Use("relationships")
	if err := db.EvalQuery(query, relationships, &queryResult); err != nil {
		return nil, err
	}
	var rels []Relationship

	for id := range queryResult {
		// To get query result document, simply read it
		readBack, err := relationships.Read(id)
		if err == nil {
			out, err := json.Marshal(readBack)
			if err == nil {
				var relationship Relationship
				json.Unmarshal(out, &relationship)
				relationship.ID = id
				rels = append(rels, relationship)
			}
		}
	}
	return rels, nil
}

// DeleteRelationship deletes a relationship given the quad.
func (cdb *CrawlDB) DeleteRelationship(onion string, from string, identiferType string, identifier string) error {
	relationships := cdb.db.Use("relationships")
	rels, err := cdb.GetRelationshipsWithOnion(onion)
	if err == nil {
		for _, rel := range rels {
			if rel.From == from && rel.Type == identiferType && rel.Identifier == identifier {
				err := relationships.Delete(rel.ID)
				return err
			}
		}
	}
	return errors.New("could not find record to delete")
}
