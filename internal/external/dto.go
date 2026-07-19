package external

type SummaryRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type SummaryResponse struct {
	SummaryID string          `json:"summary_id"`
	Schemas   []SchemaSummary `json:"schemas"`
}

type SchemaSummary struct {
	Name   string         `json:"name"`
	Tables []TableSummary `json:"tables"`
}

type TableSummary struct {
	Name     string  `json:"name"`
	RowCount int64   `json:"row_count"`
	SizeMB   float64 `json:"size_mb"`
}
