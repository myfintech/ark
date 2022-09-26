package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/tj/assert"
)

type note struct {
	Text string
}

func TestDatabase(t *testing.T) {
	cwd, _ := os.Getwd()
	testdata := filepath.Join(cwd, "testdata")
	dbfile := filepath.Join(testdata, ".testdb")
	defer os.Remove(dbfile)
	db, err := Open(dbfile)
	assert.NoError(t, err)

	err = db.Transaction(func(tx *bolt.Tx) error {
		users, _ := db.Collection("users", tx)
		user, _ := users.Collection("user_idz")
		notes, _ := user.Collection("notesz")

		assert.Equal(t, notes.Path(), []string{
			"users", "user_idz", "notesz",
		})

		err = notes.Put("reminder", &note{
			Text: "remember to do this thing",
		})
		if err != nil {
			return err
		}

		reminder := &note{}
		exists, err := notes.Get("reminder", reminder)
		if err != nil {
			return nil
		}

		assert.True(t, exists)
		assert.Equal(t, reminder.Text, "remember to do this thing")
		return nil
	})
	assert.NoError(t, err)
}
