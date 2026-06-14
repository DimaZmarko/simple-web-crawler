package crawl

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// cursor is the keyset pagination position: the (created_at, id) tuple of the
// last item on the previous page. It is serialised as an opaque base64url token
// so clients treat it as a black box.
type cursor struct {
	CreatedAt time.Time `json:"createdAt"`
	ID        uuid.UUID `json:"id"`
}

// encodeCursor serialises a cursor to an opaque base64url token.
func encodeCursor(c cursor) string {
	raw, _ := json.Marshal(c) // cursor is always serialisable; no error path.
	return base64.RawURLEncoding.EncodeToString(raw)
}

// decodeCursor parses an opaque base64url token back into a cursor. A malformed
// token yields ErrInvalidCursor so handlers can map it to a 400.
func decodeCursor(token string) (cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return cursor{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}

	var c cursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return cursor{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}

	if c.ID == uuid.Nil || c.CreatedAt.IsZero() {
		return cursor{}, fmt.Errorf("%w: missing fields", ErrInvalidCursor)
	}

	return c, nil
}
