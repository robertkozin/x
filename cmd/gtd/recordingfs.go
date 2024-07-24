package main

import (
	"bytes"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/oops"
	"io"
	"io/fs"
	"time"
)

var (
	_ fs.FS = (*recordingFS)(nil)
	_ interface {
		fs.File
		fs.FileInfo
		io.ReadSeeker
	} = (*recordingFile)(nil)
)

type recordingFS struct {
	db *pgxpool.Pool
}

func (rfs recordingFS) Open(name string) (fs.File, error) {
	errb := oops.With("name", name)
	ctx := context.Background()
	rows, _ := rfs.db.Query(ctx, `select created_at, recording from notes where filename = $1 limit 1`, name)

	exists := rows.Next()
	if !exists {
		if err := errb.Wrap(rows.Err()); err != nil {
			return nil, err
		}
		return nil, errb.Wrap(fs.ErrNotExist)
	}

	var (
		createdAt   time.Time
		driverBytes pgtype.DriverBytes
	)
	err := errb.Wrap(rows.Scan(&createdAt, &driverBytes)) // TODO: timeout?
	if err != nil {
		rows.Close()
		return nil, err
	}

	return recordingFile{
		name:     name,
		size:     int64(len(driverBytes)),
		modified: createdAt,
		rows:     rows,
		Reader:   bytes.NewReader(driverBytes),
	}, nil
}

type recordingFile struct {
	name     string
	size     int64
	modified time.Time
	rows     pgx.Rows
	*bytes.Reader
}

func (rf recordingFile) Stat() (fs.FileInfo, error) {
	return rf, nil
}

func (rf recordingFile) Close() error {
	rf.rows.Close()
	return rf.rows.Err()
}

func (rf recordingFile) Name() string {
	return rf.name
}

func (rf recordingFile) Size() int64 {
	return rf.size
}

func (rf recordingFile) Mode() fs.FileMode {
	return 0
}

func (rf recordingFile) ModTime() time.Time {
	return rf.modified
}

func (rf recordingFile) IsDir() bool {
	return false
}

func (rf recordingFile) Sys() any {
	return nil
}
