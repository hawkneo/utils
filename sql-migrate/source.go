package migrate

import (
	"os"
	"path"
)

var (
	_ MigrationSource = (*CombinedMigrationSource)(nil)
	_ MigrationSource = (*DirectoryMigrationSource)(nil)
	_ MigrationSource = (*StringMigrationSource)(nil)
)

type MigrationSource interface {
	LoadMigrations() (migrations []*Migration, err error)
}

type CombinedMigrationSource struct {
	Sources []MigrationSource
}

func (source CombinedMigrationSource) LoadMigrations() (migrations []*Migration, err error) {
	for _, ms := range source.Sources {
		subMigrations, err := ms.LoadMigrations()
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, subMigrations...)
	}
	return migrations, nil
}

type StringMigrationSource struct {
	Migrations []*Migration
}

func (source StringMigrationSource) LoadMigrations() (migrations []*Migration, err error) {
	return source.Migrations, nil
}

type DirectoryMigrationSource struct {
	Directory string
}

func (source DirectoryMigrationSource) LoadMigrations() (migrations []*Migration, err error) {
	entries, err := os.ReadDir(source.Directory)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// TODO: log
			continue
		}

		if !IsSupportFilename(entry.Name()) {
			// TODO: log
			continue
		}

		fileInfo, err := entry.Info()
		if err != nil {
			return nil, err
		}

		filePath := path.Join(source.Directory, fileInfo.Name())
		bz, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		migrations = append(migrations, &Migration{
			Filename: fileInfo.Name(),
			Source:   string(bz),
		})
	}

	return migrations, nil
}
