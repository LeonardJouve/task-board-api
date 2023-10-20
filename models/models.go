package models

import "github.com/LeonardJouve/task-board-api/store"

func AutoMigrate() error {
	if err := store.Database.AutoMigrate(&Board{}); err != nil {
		return err
	}

	if err := store.Database.AutoMigrate(&Board{}); err != nil {
		return err
	}

	if err := store.Database.AutoMigrate(&Column{}); err != nil {
		return err
	}

	if err := store.Database.AutoMigrate(&Card{}); err != nil {
		return err
	}

	if err := store.Database.AutoMigrate(&Tag{}); err != nil {
		return err
	}

	return nil
}
