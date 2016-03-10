package server

import (
	"fmt"

	"github.com/boltdb/bolt"
)

var (
	NewBucket        = "New"
	InProgressBucket = "In-Progress"
	CompletedBucket  = "Completed"
)

func InitializeBoltDb(db *bolt.DB) error {
	var err error

	err = createBoltBucket(db, NewBucket)
	if err != nil {
		return err
	}

	err = createBoltBucket(db, InProgressBucket)
	if err != nil {
		return err
	}

	err = createBoltBucket(db, CompletedBucket)
	if err != nil {
		return err
	}

	return nil
}

func createBoltBucket(db *bolt.DB, bucket string) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}
