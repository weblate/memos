package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/usememos/memos/api"
	"github.com/usememos/memos/common"
)

type storageRaw struct {
	ID        int
	Name      string
	EndPoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
	URLPrefix string
}

func (raw *storageRaw) toStorage() *api.Storage {
	return &api.Storage{
		ID:        raw.ID,
		Name:      raw.Name,
		EndPoint:  raw.EndPoint,
		Region:    raw.Region,
		AccessKey: raw.AccessKey,
		SecretKey: raw.SecretKey,
		Bucket:    raw.Bucket,
		URLPrefix: raw.URLPrefix,
	}
}

func (s *Store) CreateStorage(ctx context.Context, create *api.StorageCreate) (*api.Storage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, FormatError(err)
	}
	defer tx.Rollback()

	storageRaw, err := createStorageRaw(ctx, tx, create)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, FormatError(err)
	}

	return storageRaw.toStorage(), nil
}

func (s *Store) PatchStorage(ctx context.Context, patch *api.StoragePatch) (*api.Storage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, FormatError(err)
	}
	defer tx.Rollback()

	storageRaw, err := patchStorageRaw(ctx, tx, patch)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, FormatError(err)
	}

	return storageRaw.toStorage(), nil
}

func (s *Store) FindStorageList(ctx context.Context, find *api.StorageFind) ([]*api.Storage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, FormatError(err)
	}
	defer tx.Rollback()

	storageRawList, err := findStorageRawList(ctx, tx, find)
	if err != nil {
		return nil, err
	}

	list := []*api.Storage{}
	for _, raw := range storageRawList {
		list = append(list, raw.toStorage())
	}

	return list, nil
}

func (s *Store) FindStorage(ctx context.Context, find *api.StorageFind) (*api.Storage, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, FormatError(err)
	}
	defer tx.Rollback()

	list, err := findStorageRawList(ctx, tx, find)
	if err != nil {
		return nil, err
	}

	if len(list) == 0 {
		return nil, &common.Error{Code: common.NotFound, Err: fmt.Errorf("not found")}
	}

	storageRaw := list[0]
	return storageRaw.toStorage(), nil
}

func (s *Store) DeleteStorage(ctx context.Context, delete *api.StorageDelete) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return FormatError(err)
	}
	defer tx.Rollback()

	if err := deleteStorage(ctx, tx, delete); err != nil {
		return FormatError(err)
	}

	if err := tx.Commit(); err != nil {
		return FormatError(err)
	}

	return nil
}

func createStorageRaw(ctx context.Context, tx *sql.Tx, create *api.StorageCreate) (*storageRaw, error) {
	set := []string{"name", "end_point", "region", "access_key", "secret_key", "bucket", "url_prefix"}
	args := []interface{}{create.Name, create.EndPoint, create.Region, create.AccessKey, create.SecretKey, create.Bucket, create.URLPrefix}
	placeholder := []string{"?", "?", "?", "?", "?", "?", "?"}

	query := `
		INSERT INTO storage (
			` + strings.Join(set, ", ") + `
		)
		VALUES (` + strings.Join(placeholder, ",") + `)
		RETURNING id, name, end_point, region, access_key, secret_key, bucket, url_prefix
	`
	var storageRaw storageRaw
	if err := tx.QueryRowContext(ctx, query, args...).Scan(
		&storageRaw.ID,
		&storageRaw.Name,
		&storageRaw.EndPoint,
		&storageRaw.Region,
		&storageRaw.AccessKey,
		&storageRaw.SecretKey,
		&storageRaw.Bucket,
		&storageRaw.URLPrefix,
	); err != nil {
		return nil, FormatError(err)
	}

	return &storageRaw, nil
}

func patchStorageRaw(ctx context.Context, tx *sql.Tx, patch *api.StoragePatch) (*storageRaw, error) {
	set, args := []string{}, []interface{}{}
	if v := patch.Name; v != nil {
		set, args = append(set, "name = ?"), append(args, *v)
	}
	if v := patch.EndPoint; v != nil {
		set, args = append(set, "end_point = ?"), append(args, *v)
	}
	if v := patch.Region; v != nil {
		set, args = append(set, "region = ?"), append(args, *v)
	}
	if v := patch.AccessKey; v != nil {
		set, args = append(set, "access_key = ?"), append(args, *v)
	}
	if v := patch.SecretKey; v != nil {
		set, args = append(set, "secret_key = ?"), append(args, *v)
	}
	if v := patch.Bucket; v != nil {
		set, args = append(set, "bucket = ?"), append(args, *v)
	}
	if v := patch.URLPrefix; v != nil {
		set, args = append(set, "url_prefix = ?"), append(args, *v)
	}

	args = append(args, patch.ID)

	query := `
		UPDATE storage
		SET ` + strings.Join(set, ", ") + `
		WHERE id = ?
		RETURNING id, name, end_point, region, access_key, secret_key, bucket, url_prefix
	`

	var storageRaw storageRaw
	if err := tx.QueryRowContext(ctx, query, args...).Scan(
		&storageRaw.ID,
		&storageRaw.Name,
		&storageRaw.EndPoint,
		&storageRaw.Region,
		&storageRaw.AccessKey,
		&storageRaw.SecretKey,
		&storageRaw.Bucket,
		&storageRaw.URLPrefix,
	); err != nil {
		return nil, FormatError(err)
	}

	return &storageRaw, nil
}

func findStorageRawList(ctx context.Context, tx *sql.Tx, find *api.StorageFind) ([]*storageRaw, error) {
	where, args := []string{"1 = 1"}, []interface{}{}

	if v := find.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := find.Name; v != nil {
		where, args = append(where, "name = ?"), append(args, *v)
	}

	query := `
		SELECT
			id, 
			name, 
			end_point, 
			region,
			access_key, 
			secret_key, 
			bucket,
			url_prefix
		FROM storage
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY id DESC
	`
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, FormatError(err)
	}
	defer rows.Close()

	storageRawList := make([]*storageRaw, 0)
	for rows.Next() {
		var storageRaw storageRaw
		if err := rows.Scan(
			&storageRaw.ID,
			&storageRaw.Name,
			&storageRaw.EndPoint,
			&storageRaw.Region,
			&storageRaw.AccessKey,
			&storageRaw.SecretKey,
			&storageRaw.Bucket,
			&storageRaw.URLPrefix,
		); err != nil {
			return nil, FormatError(err)
		}

		storageRawList = append(storageRawList, &storageRaw)
	}

	if err := rows.Err(); err != nil {
		return nil, FormatError(err)
	}

	return storageRawList, nil
}

func deleteStorage(ctx context.Context, tx *sql.Tx, delete *api.StorageDelete) error {
	where, args := []string{"id = ?"}, []interface{}{delete.ID}

	stmt := `DELETE FROM storage WHERE ` + strings.Join(where, " AND ")
	result, err := tx.ExecContext(ctx, stmt, args...)
	if err != nil {
		return FormatError(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return &common.Error{Code: common.NotFound, Err: fmt.Errorf("storage not found")}
	}

	return nil
}
