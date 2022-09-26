package event_sink_server

import (
	"cloud.google.com/go/bigquery"
)

// Save implements the ValueSaver interface for inserting data into BigQuery tables
func (x *CanonicalEvent) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"org_id":          x.OrgId,
		"project_id":      x.ProjectId,
		"event_id":        x.EventId,
		"user_id":         x.UserId,
		"verb":            x.Verb,
		"method":          x.Method,
		"version":         x.Version,
		"tracker_version": x.TrackerVersion,
		"metadata":        x.Metadata,
		"created_at":      x.CreatedAt,
	}, bigquery.NoDedupeID, nil
}
