package security

import (
	"database/sql"
	"errors"
	"sync"

	"github.com/johnny1110/crypto-exchange/dto"
)

// CredentialCache maps a login token to its user. It is backed by a `credentials`
// table so a token survives a backend restart: on an in-memory miss it rehydrates
// the user from the DB. (The original was memory-only, so every restart invalidated
// every session.)
type CredentialCache struct {
	db         *sql.DB
	mu         sync.RWMutex
	tokenCache map[string]*dto.User
}

func NewCredentialCache(db *sql.DB) *CredentialCache {
	return &CredentialCache{
		db:         db,
		tokenCache: make(map[string]*dto.User),
	}
}

func (c *CredentialCache) Put(token string, user *dto.User) {
	c.mu.Lock()
	c.tokenCache[token] = user
	c.mu.Unlock()
	if c.db != nil && user != nil {
		c.db.Exec(`INSERT OR REPLACE INTO credentials(token, user_id) VALUES(?, ?)`, token, user.ID)
	}
}

func (c *CredentialCache) Delete(token string) {
	c.mu.Lock()
	delete(c.tokenCache, token)
	c.mu.Unlock()
	if c.db != nil {
		c.db.Exec(`DELETE FROM credentials WHERE token = ?`, token)
	}
}

func (c *CredentialCache) Get(token string) (*dto.User, error) {
	c.mu.RLock()
	user, ok := c.tokenCache[token]
	c.mu.RUnlock()
	if ok {
		return user, nil
	}

	// Cache miss (e.g. after a restart): rehydrate from the persisted token.
	if c.db == nil {
		return nil, errors.New("invalid credential")
	}
	var uid string
	if err := c.db.QueryRow(`SELECT user_id FROM credentials WHERE token = ?`, token).Scan(&uid); err != nil {
		return nil, errors.New("invalid credential")
	}
	u := &dto.User{}
	err := c.db.QueryRow(
		`SELECT id, username, vip_level, maker_fee, taker_fee FROM users WHERE id = ?`, uid,
	).Scan(&u.ID, &u.Username, &u.VipLevel, &u.MakerFee, &u.TakerFee)
	if err != nil {
		return nil, errors.New("invalid credential")
	}
	c.mu.Lock()
	c.tokenCache[token] = u
	c.mu.Unlock()
	return u, nil
}
