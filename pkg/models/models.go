package models

import (
	"time"
)

type Client struct {
	UID      int64     `json:"uid"`
	Birthday time.Time `json:"birthday"`
	Sex      string    `json:"sex"`
	Name     string    `json:"name"`
}

type Config struct {
	DatabaseURL string
	DataLoaded  bool
	CSVFilePath string
}
