package gormbe

import (
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type GenId struct {
	Class     string `gorm:"primaryKey"`
	Id        string `gorm:"primaryKey"`
	CreatedAt time.Time
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
