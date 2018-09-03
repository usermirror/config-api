package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/usermirror/config-api/pkg/migrations/postgres"

	// add postgres driver
	_ "github.com/lib/pq"
)

// NewPostgres creates a new Store through the lib/pq database client.
func NewPostgres(postgresAddr string) (*Postgres, error) {
	if db, err := sql.Open("postgres", postgresAddr); err != nil {
		return nil, err
	} else {
		return &Postgres{
			DB: db,
		}, nil
	}
}

// Postgres backed persistence for arbitrary key/values.
type Postgres struct {
	DB *sql.DB
}

// implements Store interface
var _ Store = new(Postgres)

func (p *Postgres) Init() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if migrationSQL, err := postgres.GetSQL("up", 0); err != nil {
		return err
	} else {
		if _, err = p.DB.ExecContext(ctx, migrationSQL); err != nil {
			return err
		}
	}

	fmt.Println("storage.postgres.init: up-to-date schema")

	return nil
}

func (p *Postgres) Get(input GetInput) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(input.Timeout)*time.Millisecond)
	defer cancel()

	var value []byte

	stmt := `SELECT value FROM namespace_configs WHERE key=$1;`
	err := p.DB.QueryRowContext(ctx, stmt, input.Key).Scan(&value)

	if err != nil {
		return nil, err
	}

	if len(value) == 0 {
		return nil, nil
	}

	return value, nil
}

func (p *Postgres) Set(input SetInput) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(input.Timeout)*time.Millisecond)
	defer cancel()

	stmt := `
		INSERT INTO namespace_configs (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE
			SET value = excluded.value;`

	_, err := p.DB.ExecContext(ctx, stmt, input.Key, input.Value)

	return err
}

func (p *Postgres) Scan(input ScanInput) (KeyList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(input.Timeout)*time.Millisecond)
	defer cancel()

	stmt := `
		SELECT key
		FROM namespace_configs
		WHERE key like ($1::bytea || '%'::bytea);`
	rows, err := p.DB.QueryContext(ctx, stmt, input.Prefix)
	if err != nil {
		fmt.Println("storage.postgres.scan: no rows found")
		return KeyList{}, err
	} else if rows == nil {
		fmt.Println("storage.postgres.scan: rows == nil")
		return KeyList{}, nil
	}

	defer rows.Close()

	kl := KeyList{}
	scanErrs := []error{}

	fmt.Println("storage.postgres.scan: got rows")

	for rows.Next() {
		var key string

		if err = rows.Scan(&key); err != nil {
			scanErrs = append(scanErrs, err)
		} else {
			kl.Keys = append(kl.Keys, strings.Replace(string(key), input.Prefix+"::", "", 1))
		}
	}

	// TODO: alert on scan errors
	if len(scanErrs) > 0 {
		fmt.Println(scanErrs)
	}

	return kl, nil
}

func (p *Postgres) Close() error {
	return p.DB.Close()
}

func (p *Postgres) CheckAuth(input AuthInput) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	namespace := input.Namespace
	stmt := `
		SELECT token
		FROM namespaces
		WHERE namespace_id=?`
	rows, err := p.DB.QueryContext(ctx, stmt, namespace)
	if err != nil {
		return fmt.Errorf("failed to execute query checking auth for namespace '%s': %v", namespace, err)
	}
	defer rows.Close()

	// token not set, allow the write
	if !rows.Next() {
		return nil
	}

	var token string
	if err = rows.Scan(&token); err != nil {
		return fmt.Errorf("could not read token for namespace '%s': %v", namespace, err)
	}

	if token != input.Token {
		return fmt.Errorf("token for namespace '%s' did not match request", namespace)
	}
	return nil
}
