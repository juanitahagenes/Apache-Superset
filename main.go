package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Column represents a dataset column
type Column struct {
	Name string
	Type string
}

// Dataset represents a database table/dataset (physical or virtual)
type Dataset struct {
	ID      string
	Name    string
	Schema  string
	SQL     string // For virtual datasets
	Columns []Column
}

// GetCacheKey returns a deterministic hash representing the dataset's schema/metadata
func (d *Dataset) GetCacheKey() string {
	h := sha256.New()
	// Include schema and name
	h.Write([]byte(d.Schema + "." + d.Name + "\n"))
	// Include SQL query if virtual
	h.Write([]byte(d.SQL + "\n"))
	// Include columns (names and types) in order
	for _, col := range d.Columns {
		h.Write([]byte(col.Name + ":" + col.Type + "\n"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// QueryObject represents a query run against a dataset
type QueryObject struct {
	Dataset   *Dataset
	QuerySQL  string
	Extras    map[string]interface{}
}

// CacheKey generates a unique cache key for the query, incorporating the dataset's cache key
func (q *QueryObject) CacheKey() string {
	h := sha256.New()
	// Incorporate the dataset's schema/metadata cache key
	h.Write([]byte(q.Dataset.GetCacheKey() + "\n"))
	// Incorporate the query details
	h.Write([]byte(q.QuerySQL + "\n"))
	return hex.EncodeToString(h.Sum(nil))
}

// CacheManager handles cache operations
type CacheManager struct {
	store map[string]string
}

func NewCacheManager() *CacheManager {
	return &CacheManager{store: make(map[string]string)}
}

func (cm *CacheManager) Get(key string) (string, bool) {
	val, ok := cm.store[key]
	return val, ok
}

func (cm *CacheManager) Set(key string, val string) {
	cm.store[key] = val
}

// InvalidateDatasetCache explicitly invalidates all cache entries for a dataset ID
func (cm *CacheManager) InvalidateDatasetCache(datasetID string) {
	for key := range cm.store {
		if strings.Contains(key, datasetID) {
			delete(cm.store, key)
		}
	}
}

func main() {
	fmt.Println("Starting Cache Invalidation Demo...")

	// 1. Create a dataset
	dataset := &Dataset{
		ID:     "dataset_1",
		Name:   "sales_data",
		Schema: "public",
		Columns: []Column{
			{Name: "id", Type: "INTEGER"},
			{Name: "amount", Type: "DECIMAL"},
		},
	}

	// 2. Create a query object
	query := &QueryObject{
		Dataset:  dataset,
		QuerySQL: "SELECT id, amount FROM public.sales_data",
	}

	// Generate initial cache key
	key1 := query.CacheKey()
	fmt.Printf("Initial Cache Key: %s\n", key1)

	// Store result in cache
	cache := NewCacheManager()
	cache.Set(key1, "cached_query_results_v1")

	// Verify cache hit
	if val, found := cache.Get(key1); found {
		fmt.Printf("Cache Hit: %s\n", val)
	} else {
		fmt.Println("Cache Miss!")
	}

	// 3. Update dataset schema (e.g., add a new column)
	fmt.Println("\nUpdating dataset schema (adding 'discount' column)...")
	dataset.Columns = append(dataset.Columns, Column{Name: "discount", Type: "DECIMAL"})

	// Generate new cache key
	key2 := query.CacheKey()
	fmt.Printf("New Cache Key:     %s\n", key2)

	// Verify cache miss due to changed cache key (Implicit Invalidation)
	if key1 == key2 {
		fmt.Println("Error: Cache key did not change after schema update!")
	} else {
		fmt.Println("Success: Cache key changed successfully (Implicit Invalidation).")
	}

	if _, found := cache.Get(key2); !found {
		fmt.Println("Cache Miss on new key (as expected).")
	} else {
		fmt.Println("Error: Cache hit on new key!")
	}
}
