package gormbe

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type GenId struct {
	Class      string `gorm:"primaryKey"`
	Id         string `gorm:"primaryKey"`
	CreatedAt  time.Time
	VerifiedAt time.Time
	Released   bool
}

func randid() string {
	max_id := int64(math.Pow(36, 8))
	randval := rand.Int63() % max_id
	return strconv.FormatInt(randval, 36)
}

// Generate 1 New ID
func NewID(storage *gorm.DB, cls string) string {
	for {
		gid := GenId{Id: randid(), Class: cls, CreatedAt: time.Now()}
		err := storage.Create(gid).Error
		if err == nil {
			return gid.Id
		} else {
			log.Println("ID Create Error: ", err)
		}
	}
}

/**
 * Create N IDs in batch.
 */
func NewIDs(storage *gorm.DB, cls string, numids int) (out []string) {
	for i := 0; i < numids; i++ {
		for {
			gid := GenId{Id: randid(), Class: cls, CreatedAt: time.Now()}
			err := storage.Create(gid).Error
			if err != nil {
				log.Println("ID Create Error: ", i, err)
			} else {
				out = append(out, gid.Id)
				break
			}
		}
	}
	return
}

func VerifyID(storage *gorm.DB, cls string, id string) error {
	var gid GenId
	err := storage.First(&gid, "cls = ? and id = ?", cls, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	gid.VerifiedAt = time.Now()
	return storage.Updates(gid).Error
}

func ReleaseID(storage *gorm.DB, cls string, id string) error {
	var gid GenId
	err := storage.First(&gid, "cls = ? and id = ?", cls, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	gid.Released = true
	gid.VerifiedAt = time.Now()
	return storage.Updates(gid).Error
}
