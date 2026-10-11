package repositories

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// pageCursor identifies the last item of a page. Listings are sorted by createdAt desc
// with id as tiebreaker, so both are needed to resume from the right position.
type pageCursor struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        string    `json:"id"`
}

func encodeCursor(cursor pageCursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeCursor(encoded string) (pageCursor, error) {
	var cursor pageCursor
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return cursor, err
	}
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return cursor, err
	}
	if cursor.ID == "" || cursor.CreatedAt.IsZero() {
		return cursor, errors.New("cursor is missing fields")
	}
	return cursor, nil
}
