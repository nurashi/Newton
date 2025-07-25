package db

import "context"

func InitSchema() error {
	schema := ``

	_, err := Pool.Exec(context.Background(), schema)

	return err
}