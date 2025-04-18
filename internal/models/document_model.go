package models

import (
	"fmt"
	"strings"
	"time"
)

type DocumentType string

const (
	DocumentTypePassport      DocumentType = "passport"
	DocumentTypeIDCard        DocumentType = "id_card"
	DocumentTypeDriverLicense DocumentType = "driver_licence"
	DocumentTypeOther         DocumentType = "other"
)

type Document struct {
	ID          string       `json:"id" firestore:"id"`
	UserID      string       `json:"user_id" firestore:"user_id" `
	Name        string       `json:"name" firestore:"name"`
	Size        int64        `json:"size" firestore:"size"`
	Type        DocumentType `json:"type" firestore:"type"`
	ContentType string       `json:"content_type" firestore:"content_type"`
	Path        string       `json:"path" firestore:"path"`
	Bucket      string       `json:"bucket" firestore:"bucket"`
	CreatedAt   time.Time    `json:"created_at" firestore:"created_at,serverTimestamp"`
	UpdatedAt   time.Time    `json:"updated_at" firestore:"updated_at,serverTimestamp"`
}

func ParseDocumentType(docType string) (DocumentType, error) {
	switch strings.ToUpper(docType) {
	case "PASSPORT":
		return DocumentTypePassport, nil
	case "ID_CARD":
		return DocumentTypeIDCard, nil
	case "DRIVER_LICENSE":
		return DocumentTypeDriverLicense, nil
	// Add other types as needed
	default:
		return "", fmt.Errorf("unknown document type: %s", docType)
	}
}
