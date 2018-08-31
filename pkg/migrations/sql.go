package migrations

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// GetSQL connects to a database then runs a migration
// based on the direction (e.g. up/down) and schema version
func GetSQL(dbType string, direction string, version int) (string, error) {
	files, err := ioutil.ReadDir(fmt.Sprintf("./%s", dbType))
	if err != nil {
		return "", err
	}

	var fullSQL string

	for _, file := range files {
		fileName := file.Name()

		if strings.Contains(fileName, direction) {
			nameParts := strings.Split(fileName, "_")
			schemaVersion, err := strconv.Atoi(nameParts[0])
			if err != nil {
				return "", err
			}

			if version < schemaVersion {
				sqlBytes, err := ioutil.ReadFile(fmt.Sprintf("./%s/%s", dbType, fileName))
				if err != nil {
					return "", err
				}

				fullSQL += "\n" + string(sqlBytes)
			}
		}
	}

	return fullSQL, nil
}
