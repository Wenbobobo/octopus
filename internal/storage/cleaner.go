package storage

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/duo/octopus/internal/common"
	"github.com/duo/octopus/internal/db"
	"github.com/duo/octopus/internal/manager"

	log "github.com/sirupsen/logrus"
)

type Cleaner struct {
	config *common.Configure

	stop chan struct{}
	done chan struct{}
}

func NewCleaner(config *common.Configure) *Cleaner {
	return &Cleaner{
		config: config,
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}
}

func (c *Cleaner) enabled() bool {
	storage := c.config.Service.Storage
	return storage.MaxTotalBytes > 0 || storage.MessageTTLDays > 0
}

func (c *Cleaner) Start() {
	if !c.enabled() {
		return
	}

	go func() {
		defer close(c.done)

		c.runOnce("startup")
		ticker := time.NewTicker(c.config.Service.Storage.CleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.runOnce("ticker")
			case <-c.stop:
				return
			}
		}
	}()
}

func (c *Cleaner) Stop() {
	if !c.enabled() {
		return
	}
	close(c.stop)
	<-c.done
}

func (c *Cleaner) runOnce(trigger string) {
	storage := c.config.Service.Storage
	vacuumNeeded := false

	if storage.MessageTTLDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -storage.MessageTTLDays).Unix()
		deleted, err := manager.DeleteMessagesOlderThan(cutoff)
		if err != nil {
			log.Warnf("StorageCleaner(%s): delete messages by ttl failed: %v", trigger, err)
		} else if deleted > 0 {
			vacuumNeeded = true
			log.Infof("StorageCleaner(%s): deleted %d expired messages", trigger, deleted)
		}
	}

	if storage.MaxTotalBytes <= 0 {
		if vacuumNeeded {
			c.vacuum("ttl")
		}
		return
	}

	size, err := dirSize(storage.DataDir)
	if err != nil {
		log.Warnf("StorageCleaner(%s): calc dir size failed: %v", trigger, err)
		return
	}

	if size <= storage.MaxTotalBytes {
		if vacuumNeeded {
			c.vacuum("ttl")
		}
		return
	}

	target := storage.TargetTotalBytes
	if target <= 0 || target >= storage.MaxTotalBytes {
		target = storage.MaxTotalBytes * 7 / 10
	}

	log.Warnf("StorageCleaner(%s): data dir is over limit, size=%d max=%d target=%d", trigger, size, storage.MaxTotalBytes, target)

	for size > target {
		deleted, err := manager.DeleteOldestMessages(storage.BatchDelete)
		if err != nil {
			log.Warnf("StorageCleaner(%s): delete oldest messages failed: %v", trigger, err)
			break
		}
		if deleted == 0 {
			break
		}
		vacuumNeeded = true

		size, err = dirSize(storage.DataDir)
		if err != nil {
			log.Warnf("StorageCleaner(%s): recalc dir size failed: %v", trigger, err)
			break
		}
	}

	if vacuumNeeded {
		c.vacuum("size")
	}
}

func (c *Cleaner) vacuum(reason string) {
	if _, err := db.DB.Exec(`VACUUM;`); err != nil {
		log.Warnf("StorageCleaner(%s): vacuum failed: %v", reason, err)
	} else {
		count, err := manager.CountMessages()
		if err != nil {
			log.Infof("StorageCleaner(%s): vacuum done", reason)
		} else {
			log.Infof("StorageCleaner(%s): vacuum done, message rows=%d", reason, count)
		}
	}
}

func dirSize(root string) (int64, error) {
	var size int64

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		size += info.Size()
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	return size, nil
}
