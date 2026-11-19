package listfilter

import (
	"strconv"
	"strings"

	"gorm.io/gorm"
)

func Search(tx *gorm.DB, search, idColumn string, textColumns ...string) *gorm.DB {
	search = strings.TrimSpace(search)
	if search == "" {
		return tx
	}
	pattern := "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(strings.ToLower(search)) + "%"
	conditions := make([]string, 0, len(textColumns)+1)
	args := make([]interface{}, 0, len(textColumns)+1)
	for _, column := range textColumns {
		conditions = append(conditions, "LOWER("+column+") LIKE ? ESCAPE '!'")
		args = append(args, pattern)
	}
	if id, err := strconv.ParseUint(search, 10, 64); err == nil && id > 0 && idColumn != "" {
		conditions = append(conditions, idColumn+" = ?")
		args = append(args, id)
	}
	if len(conditions) == 0 {
		return tx
	}
	return tx.Where("("+strings.Join(conditions, " OR ")+")", args...)
}

func Pagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 5000 {
		limit = 5000
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
