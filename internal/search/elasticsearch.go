package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ElasticsearchClient wraps the Elasticsearch client
type ElasticsearchClient struct {
	client *elasticsearch.Client
	index  string
}

// Config holds Elasticsearch configuration
type Config struct {
	Addresses []string
	Index     string
}

// DetectionDocument represents an indexed detection event
type DetectionDocument struct {
	EventID      int                    `json:"event_id"`
	FrameID      *int                   `json:"frame_id,omitempty"`
	CameraID     int                    `json:"camera_id"`
	Timestamp    time.Time              `json:"timestamp"`
	EventType    string                 `json:"event_type"`
	Severity     string                 `json:"severity"`
	Title        string                 `json:"title"`
	Message      string                 `json:"message"`
	FramePath    string                 `json:"frame_path,omitempty"`
	HasFrame     bool                   `json:"has_frame"`
	Detections   []Detection            `json:"detections,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Detection represents a detected object
type Detection struct {
	ObjectClass string  `json:"object_class"`
	Confidence  float64 `json:"confidence"`
	BBox        BBox    `json:"bbox,omitempty"`
}

// BBox represents bounding box coordinates
type BBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// NewElasticsearchClient creates a new Elasticsearch client
func NewElasticsearchClient(cfg *Config) (*ElasticsearchClient, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch returned error: %s", res.Status())
	}

	log.Println("✅ Connected to Elasticsearch")

	esClient := &ElasticsearchClient{
		client: client,
		index:  cfg.Index,
	}

	// Create index if it doesn't exist
	if err := esClient.createIndex(); err != nil {
		return nil, err
	}

	return esClient, nil
}

// createIndex creates the index with mapping if it doesn't exist
func (es *ElasticsearchClient) createIndex() error {
	// Check if index exists
	res, err := es.client.Indices.Exists([]string{es.index})
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		log.Printf("✅ Index '%s' already exists", es.index)
		return nil
	}

	// Create index with mapping
	mapping := `{
		"mappings": {
			"properties": {
				"event_id": { "type": "integer" },
				"frame_id": { "type": "integer" },
				"camera_id": { "type": "integer" },
				"timestamp": { "type": "date" },
				"event_type": { "type": "keyword" },
				"severity": { "type": "keyword" },
				"title": { "type": "text" },
				"message": {
					"type": "text",
					"fields": {
						"keyword": { "type": "keyword" }
					}
				},
				"frame_path": { "type": "keyword" },
				"has_frame": { "type": "boolean" },
				"detections": {
					"type": "nested",
					"properties": {
						"object_class": { "type": "keyword" },
						"confidence": { "type": "float" },
						"bbox": {
							"properties": {
								"x": { "type": "integer" },
								"y": { "type": "integer" },
								"width": { "type": "integer" },
								"height": { "type": "integer" }
							}
						}
					}
				}
			}
		}
	}`

	res, err = es.client.Indices.Create(
		es.index,
		es.client.Indices.Create.WithBody(bytes.NewReader([]byte(mapping))),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.Status())
	}

	log.Printf("✅ Created index '%s'", es.index)
	return nil
}

// IndexDocument indexes a document (accepts any map or struct)
func (es *ElasticsearchClient) IndexDocument(ctx context.Context, doc interface{}) error {
	// Convert to DetectionDocument if it's a map
	var detectionDoc *DetectionDocument

	if docMap, ok := doc.(map[string]interface{}); ok {
		// Convert map to DetectionDocument
		jsonData, err := json.Marshal(docMap)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %w", err)
		}

		detectionDoc = &DetectionDocument{}
		if err := json.Unmarshal(jsonData, detectionDoc); err != nil {
			return fmt.Errorf("failed to unmarshal document: %w", err)
		}
	} else if dd, ok := doc.(*DetectionDocument); ok {
		detectionDoc = dd
	} else {
		return fmt.Errorf("unsupported document type: %T", doc)
	}

	return es.indexDetectionDocument(ctx, detectionDoc)
}

// indexDetectionDocument indexes a detection document
func (es *ElasticsearchClient) indexDetectionDocument(ctx context.Context, doc *DetectionDocument) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      es.index,
		DocumentID: fmt.Sprintf("event_%d", doc.EventID),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index document: %s", res.Status())
	}

	return nil
}

// BulkIndexDocuments indexes multiple documents in bulk
func (es *ElasticsearchClient) BulkIndexDocuments(ctx context.Context, docs []*DetectionDocument) error {
	if len(docs) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, doc := range docs {
		// Action metadata
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": es.index,
				"_id":    fmt.Sprintf("event_%d", doc.EventID),
			},
		}
		metaJSON, _ := json.Marshal(meta)
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		// Document
		docJSON, _ := json.Marshal(doc)
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	res, err := es.client.Bulk(bytes.NewReader(buf.Bytes()), es.client.Bulk.WithIndex(es.index))
	if err != nil {
		return fmt.Errorf("bulk index failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index error: %s", res.Status())
	}

	var bulkRes map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&bulkRes); err != nil {
		return fmt.Errorf("failed to parse bulk response: %w", err)
	}

	if bulkRes["errors"].(bool) {
		log.Printf("⚠️  Some documents failed to index")
	}

	log.Printf("✅ Bulk indexed %d documents", len(docs))
	return nil
}

// Search performs a search query
func (es *ElasticsearchClient) Search(ctx context.Context, query map[string]interface{}) ([]DetectionDocument, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	res, err := es.client.Search(
		es.client.Search.WithContext(ctx),
		es.client.Search.WithIndex(es.index),
		es.client.Search.WithBody(&buf),
		es.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	docs := make([]DetectionDocument, 0, len(hits))

	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"]
		docJSON, _ := json.Marshal(source)
		var doc DetectionDocument
		json.Unmarshal(docJSON, &doc)
		docs = append(docs, doc)
	}

	return docs, nil
}

// Close closes the Elasticsearch client
func (es *ElasticsearchClient) Close() error {
	// Elasticsearch v8 client doesn't require explicit close
	return nil
}
