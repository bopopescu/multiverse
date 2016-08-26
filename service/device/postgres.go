package device

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/tapglue/multiverse/platform/flake"
	"github.com/tapglue/multiverse/platform/pg"
)

const (
	pgInsertDevice = `INSERT INTO
		%s.devices(deleted, device_id, disabled, endpoint_arn, id, language, platform, token, user_id, created_at, updated_at)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	pgUpdateDevice = `
		UPDATE
			%s.devices
		SET
			deleted = $2,
			disabled = $3,
			endpoint_arn = $4,
			language = $5,
			token = $6,
			updated_at = $7
		WHERE
			id = $1`

	pgListDevices = `
		SELECT
			deleted, device_id, disabled, endpoint_arn, id, language, platform, token, user_id, created_at, updated_at
		FROM
			%s.devices
		%s`

	pgClauseDeleted      = `deleted = ?`
	pgClauseDeviceIDs    = `device_id IN (?)`
	pgClauseDisabled     = `disabled = ?`
	pgClauseEndpointARNs = `endpoint_arn IN (?)`
	pgClauseIDs          = `id IN (?)`
	pgClausePlatforms    = `platform IN (?)`
	pgClauseUserIDs      = `user_id IN (?)`

	pgOrderCreatedAt = `ORDER BY created_at DESC`

	pgIndexDeviceIDUserID = `
		CREATE INDEX
			%s
		ON
			%s.devices (device_id, user_id)
		WHERE
			deleted = false`
	pgIndexEndpointARN = `
		CREATE INDEX
			%s
		ON
			%s.devices(endpoint_arn)
		WHERE
			deleted = false`
	pgIndexID          = `CREATE INDEX %s ON %s.devices (id)`
	pgIndexUserDevices = `
		CREATE INDEX
			%s
		ON
			%s.devices(user_id)
		WHERE
			deleted = false
			AND disabled = false
			AND platform IN (1, 2, 3)`

	pgCreateSchema = `CREATE SCHEMA IF NOT EXISTS %s`
	pgCreateTable  = `CREATE TABLE IF NOT EXISTS %s.devices (
		deleted BOOL DEFAULT false,
		device_id TEXT NOT NULL,
		disabled BOOL DEFAULT false,
		endpoint_arn TEXT,
		id BIGINT NOT NULL,
		language TEXT NOT NULL,
		platform INT NOT NULL,
		token TEXT NOT NULL,
		user_id BIGINT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	)`
	pgDropTable = `DROP TABLE IF EXISTS %s.devices`
)

type pgService struct {
	db *sqlx.DB
}

// PostgresService returns a Postgres based Service implementation.
func PostgresService(db *sqlx.DB) Service {
	return &pgService{
		db: db,
	}
}

func (s *pgService) Put(ns string, d *Device) (*Device, error) {
	var (
		params []interface{}
		query  string
	)

	if err := d.Validate(); err != nil {
		return nil, err
	}

	if d.ID == 0 {
		if d.CreatedAt.IsZero() {
			d.CreatedAt = time.Now().UTC()
		}

		ts, err := time.Parse(pg.TimeFormat, d.CreatedAt.UTC().Format(pg.TimeFormat))
		if err != nil {
			return nil, err
		}

		d.CreatedAt = ts
		d.UpdatedAt = ts

		id, err := flake.NextID(flakeNamespace(ns))
		if err != nil {
			return nil, err
		}

		d.ID = id

		params = []interface{}{
			d.Deleted,
			d.DeviceID,
			d.Disabled,
			d.EndpointARN,
			d.ID,
			d.Language,
			d.Platform,
			d.Token,
			d.UserID,
			ts,
			ts,
		}
		query = fmt.Sprintf(pgInsertDevice, ns)
	} else {
		now, err := time.Parse(pg.TimeFormat, time.Now().UTC().Format(pg.TimeFormat))
		if err != nil {
			return nil, err
		}

		d.UpdatedAt = now

		params = []interface{}{
			d.ID,
			d.Deleted,
			d.Disabled,
			d.EndpointARN,
			d.Language,
			d.Token,
			d.UpdatedAt,
		}
		query = fmt.Sprintf(pgUpdateDevice, ns)
	}

	_, err := s.db.Exec(query, params...)
	if err != nil {
		if pg.IsRelationNotFound(pg.WrapError(err)) {
			if err := s.Setup(ns); err != nil {
				return nil, err
			}

			_, err = s.db.Exec(query, params...)
		}
	}

	return d, err
}

func (s *pgService) Query(ns string, opts QueryOptions) (List, error) {
	clauses, params, err := convertOpts(opts)
	if err != nil {
		return nil, err
	}

	ds, err := s.listDevices(ns, clauses, params...)
	if err != nil {
		if pg.IsRelationNotFound(pg.WrapError(err)) {
			if err := s.Setup(ns); err != nil {
				return nil, err
			}
		}

		ds, err = s.listDevices(ns, clauses, params...)
	}

	return ds, err
}

func (s *pgService) listDevices(
	ns string,
	clauses []string,
	params ...interface{},
) (List, error) {
	c := strings.Join(clauses, "\nAND ")

	if len(clauses) > 0 {
		c = fmt.Sprintf("WHERE %s", c)
	}

	query := strings.Join([]string{
		fmt.Sprintf(pgListDevices, ns, c),
		pgOrderCreatedAt,
	}, "\n")

	query = sqlx.Rebind(sqlx.DOLLAR, query)

	rows, err := s.db.Query(query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ds := List{}

	for rows.Next() {
		d := &Device{}

		err := rows.Scan(
			&d.Deleted,
			&d.DeviceID,
			&d.Disabled,
			&d.EndpointARN,
			&d.ID,
			&d.Language,
			&d.Platform,
			&d.Token,
			&d.UserID,
			&d.CreatedAt,
			&d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		d.CreatedAt = d.CreatedAt.UTC()
		d.UpdatedAt = d.UpdatedAt.UTC()

		ds = append(ds, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ds, nil
}

func (s *pgService) Setup(ns string) error {
	qs := []string{
		fmt.Sprintf(pgCreateSchema, ns),
		fmt.Sprintf(pgCreateTable, ns),
		pg.GuardIndex(ns, "device_device_id_user_id", pgIndexDeviceIDUserID),
		pg.GuardIndex(ns, "device_endpoint_arn", pgIndexEndpointARN),
		pg.GuardIndex(ns, "device_id", pgIndexID),
		pg.GuardIndex(ns, "device_user_devices", pgIndexUserDevices),
	}

	for _, q := range qs {
		_, err := s.db.Exec(q)
		if err != nil {
			return fmt.Errorf("setup (%s): %s", q, err)
		}
	}

	return nil
}

func (s *pgService) Teardown(ns string) error {
	qs := []string{
		fmt.Sprintf(pgDropTable, ns),
	}

	for _, q := range qs {
		_, err := s.db.Exec(q)
		if err != nil {
			return fmt.Errorf("teardown (%s): %s", q, err)
		}
	}

	return nil
}

func convertOpts(opts QueryOptions) ([]string, []interface{}, error) {
	var (
		clauses = []string{}
		params  = []interface{}{}
	)

	if opts.Deleted != nil {
		clause, _, err := sqlx.In(pgClauseDeleted, []interface{}{*opts.Deleted})
		if err != nil {
			return nil, nil, err
		}

		clauses = append(clauses, clause)
		params = append(params, *opts.Deleted)
	}

	if len(opts.DeviceIDs) > 0 {
		ps := []interface{}{}

		for _, id := range opts.DeviceIDs {
			ps = append(ps, id)
		}

		clause, _, err := sqlx.In(pgClauseDeviceIDs, ps)
		if err != nil {
			return nil, nil, err
		}

		clauses = append(clauses, clause)
		params = append(params, ps...)
	}

	if opts.Disabled != nil {
		clause, _, err := sqlx.In(pgClauseDisabled, []interface{}{*opts.Disabled})
		if err != nil {
			return nil, nil, err
		}

		clauses = append(clauses, clause)
		params = append(params, *opts.Disabled)
	}

	if len(opts.EndpointARNs) > 0 {
		ps := []interface{}{}

		for _, arn := range opts.EndpointARNs {
			ps = append(ps, arn)
		}

		clause, _, err := sqlx.In(pgClauseEndpointARNs, ps)
		if err != nil {
			return nil, nil, err
		}

		clauses = append(clauses, clause)
		params = append(params, ps...)
	}

	if len(opts.IDs) > 0 {
		ps := []interface{}{}

		for _, id := range opts.IDs {
			ps = append(ps, id)
		}

		clause, _, err := sqlx.In(pgClauseIDs, ps)
		if err != nil {
			return nil, nil, err
		}

		clauses = append(clauses, clause)
		params = append(params, ps...)
	}

	if len(opts.Platforms) > 0 {
		ps := []interface{}{}

		for _, p := range opts.Platforms {
			ps = append(ps, p)
		}

		clause, _, err := sqlx.In(pgClausePlatforms, ps)
		if err != nil {
			return nil, nil, err
		}

		clauses = append(clauses, clause)
		params = append(params, ps...)
	}

	if len(opts.UserIDs) > 0 {
		ps := []interface{}{}

		for _, id := range opts.UserIDs {
			ps = append(ps, id)
		}

		clause, _, err := sqlx.In(pgClauseUserIDs, ps)
		if err != nil {
			return nil, nil, err
		}

		clauses = append(clauses, clause)
		params = append(params, ps...)
	}

	return clauses, params, nil
}
