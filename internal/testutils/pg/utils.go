package pg

import (
	"fmt"
	"gorm.io/gorm"
	"ralts-cms/internal/machines"
	"ralts-cms/internal/maintenance"
)

func TruncateTables(db *gorm.DB) error {
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&machines.Machine{})
	if res.Error != nil {
		return fmt.Errorf("failed to truncate Machine table %e", res.Error)
	}

	res = db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&maintenance.Maintenance{})
	if res.Error != nil {
		return fmt.Errorf("failed to truncate Maintenance table %e", res.Error)
	}

	return nil
}
