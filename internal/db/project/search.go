package project

import (
	"github.com/jinzhu/gorm"
)

type searchResult struct {
	ProjectID int64
}

// Search - find project by contract hash
func Search(db *gorm.DB, hash string) (p Project, err error) {
	var result searchResult
	if err = db.Raw(`SELECT R.p_id AS project_id FROM (
		SELECT similarity(C.hash_code, ?), C.project_id as p_id
		FROM   contracts AS C
		WHERE C.hash_code % ?
		ORDER BY C.id asc) as R limit 1;`,
		hash, hash).Scan(&result).Error; gorm.IsRecordNotFoundError(err) {
		err = db.Create(&p).Error
		return
	} else if err != nil {
		return
	}
	return Project{
		ID: result.ProjectID,
	}, nil
}
