package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hawkneo/utils/log"
	"github.com/hawkneo/utils/sql-migrate"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	// SchemaName is the name of the schema that will be used to store the migration history.
	// If not set, the default schema name is "migration_schema".
	SchemaName string `json:"schema_name"`
	// Dialect is the database dialect used to generate the SQL statements.
	//
	// Support dialects are: postgres, mysql.
	//
	// If not set, the default dialect is "postgres".
	Dialect string `json:"dialect"`
	// DataSourceName is the database connection string.
	// See sql.Open for details.
	DataSourceName string `json:"data_source_name"`
	// WorkingDirectory is the working directory.
	WorkingDirectory string `json:"working_directory"`

	// MigrateOutOfOrder is a flag that if you already have versions 1.0 and 3.0 applied, and now a version 2.0 is found, it will be applied too instead of being ignored.
	MigrateOutOfOrder bool `json:"migrate_out_of_order"`
	// DisableColorOutput is a flag that disables colored output.
	DisableColorOutput bool `json:"disable_color_output"`
	// LoggerLevel is the level of the logger.
	LoggerLevel string `json:"logger_level"`

	// MigrationSource is the directory containing the migration files.
	// 	If not set, the default directory is "migrations".
	//
	// The migration files must be named in the following format:
	// 	V<version>__<description>.sql, where <version> is the version number and <description> is a short description of the migration.
	// 	The version number can be any positive number or text, but it must be unique.
	// 	The description can be any text. It is recommended to use a short description of the migration.
	// 	For example:
	//	- V1__Create_users_table.sql
	//	- V2__Add_email_column.sql
	//	- V2.1__Add_email_index.sql
	//	- V2.2__Add_email_unique_constraint.sql
	//	- V3__Add_password_column.sql
	MigrationSource string `json:"migration_source"`
}

func main() {
	rootCmd := &cobra.Command{
		Use: "migrate",
	}

	configPath := rootCmd.PersistentFlags().StringP("config", "c", "migration_config.json", "config file in json format")
	workingDirectory := rootCmd.PersistentFlags().StringP("working-dir", "w", "", "working directory")
	var migrateCtx *migrate.Context
	preRunE := func(cmd *cobra.Command, args []string) error {
		conf, err := readConfig(workingDirectory, configPath)
		if err != nil {
			return err
		}
		migrateConf, err := convertToMigrateConfig(conf)
		if err != nil {
			return err
		}

		migrateCtx = &migrate.Context{
			Context: context.Background(),
			Conf:    migrateConf,
		}
		return nil
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:     "create",
		Short:   "Create a migration schema in the database",
		PreRunE: preRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			migrator, err := migrate.NewCreateMigrator(migrateCtx)
			if err != nil {
				return err
			}

			return migrator.Apply()
		},
	}, &cobra.Command{
		Use:     "baseline",
		Short:   "Baselines an existing database",
		PreRunE: preRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			migrator, err := migrate.NewBaselineMigrator(migrateCtx)
			if err != nil {
				return err
			}

			return migrator.Apply()
		},
	}, &cobra.Command{
		Use:     "status",
		Short:   "Show the status of the migrations",
		PreRunE: preRunE,
		RunE: func(cmd *cobra.Command, args []string) error {
			migrator, err := migrate.NewStatusMigrator(migrateCtx)
			if err != nil {
				return err
			}

			return migrator.Apply()
		},
	})

	bindDryRunFlagFn := func(cmd *cobra.Command) *bool {
		return cmd.Flags().Bool("dry-run", false, "Print the SQL statements that will be executed without executing them")
	}

	{
		upCmd := &cobra.Command{
			Use:     "up",
			Short:   "Perform a migration to the latest version",
			PreRunE: preRunE,
		}
		dryRunFlag := bindDryRunFlagFn(upCmd)
		upCmd.RunE = func(cmd *cobra.Command, args []string) error {
			migrator, err := migrate.NewUpMigrator(migrateCtx)
			if err != nil {
				return err
			}
			migrateCtx.Conf.DryRun = *dryRunFlag

			return migrator.Apply()
		}
		rootCmd.AddCommand(upCmd)
	}

	{
		downCmd := &cobra.Command{
			Use:   "down <to-version>",
			Short: "Perform a migration to the specified version (not including this version)",
			Long: strings.TrimSpace(fmt.Sprintf(`
Example:
  To migrate down to version 1:
  $ %s down 1
  To migrate down to first verison:
  $ %s down --all
`, os.Args[0], os.Args[0],
			)),
			Args:    cobra.MaximumNArgs(1),
			PreRunE: preRunE,
		}
		dryRunFlag := bindDryRunFlagFn(downCmd)
		downAllFlag := downCmd.Flags().Bool("all", false, "Perform a migration to the first version (include this version)")
		downCmd.RunE = func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !*downAllFlag {
				return fmt.Errorf("to-version must be set")
			}
			toVersion := ""
			if len(args) == 1 {
				toVersion = args[0]
			}
			migrator, err := migrate.NewDownMigrator(migrateCtx, toVersion)
			migrateCtx.Conf.DryRun = *dryRunFlag
			if err != nil {
				return err
			}

			return migrator.Apply()
		}
		rootCmd.AddCommand(downCmd)
	}

	newCmd := &cobra.Command{
		Use:   "new",
		Short: "Create a new migration file or a new migration config file",
	}
	rootCmd.AddCommand(newCmd)

	var newVersion string
	newVersionCmd := &cobra.Command{
		Use:   "version [description]",
		Short: "Create a new migration file with description",
		Long: strings.TrimSpace(fmt.Sprintf(`
Example:
  Create a new migration file with description:
  $ %s new version "create table"
  Create a new migration file with description and version:
  $ %s new version "create table" -v %s
`, os.Args[0], os.Args[0], time.Now().Format("20060102150405"),
		)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := readConfig(workingDirectory, configPath)
			if err != nil {
				return err
			}
			migrationSource := path.Join(conf.WorkingDirectory, conf.MigrationSource)
			if _, err := os.Stat(migrationSource); os.IsNotExist(err) {
				if err = os.Mkdir(migrationSource, 0666); err != nil {
					return fmt.Errorf("failed to create migration source directory: %w", err)
				}
			}

			if newVersion == "" {
				newVersion = time.Now().Format("20060102150405")
			}
			description := strings.ReplaceAll(strings.TrimSpace(args[0]), " ", "_")
			filename := fmt.Sprintf("V%s__%s.sql", newVersion, description)

			file, err := os.OpenFile(filepath.Join(migrationSource, filename), os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				return fmt.Errorf("failed to create migration file: %w", err)
			}
			defer file.Close()

			_, err = file.WriteString(strings.TrimSpace(`
-- +migrate Up

-- +migrate Down

`))
			if err != nil {
				return fmt.Errorf("failed to write migration file: %w", err)
			}

			logger := log.AnsiColorLogger{ColorOutput: !conf.DisableColorOutput, Level: log.LevelFromString(conf.LoggerLevel)}

			logger.Infof("Created migration file: %s success", filepath.Join(conf.MigrationSource, filename))
			return nil
		},
	}
	newVersionCmd.Flags().StringVarP(&newVersion, "version", "v", "",
		"The version number can be any positive number or text, but it must be unique.",
	)

	newCmd.AddCommand(newVersionCmd)

	newCmd.AddCommand(&cobra.Command{
		Use:   "config <filename>",
		Short: "Create a new migration config file",
		Long: strings.TrimSpace(fmt.Sprintf(`
Example:
  Create a new migration config file with filename:
  $ %s new config custom_migration_config.json
  Create a new migration config file with default filename:
  $ %s new config
`, os.Args[0], os.Args[0],
		)),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := "migration_config.json"
			if len(args) == 1 {
				filename = args[0]
			}
			file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				return fmt.Errorf("failed to create migration config file: %w", err)
			}
			defer file.Close()

			config := Config{
				SchemaName:         "migration_schema",
				Dialect:            "postgres",
				DataSourceName:     "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable TimeZone=Etc/UTC",
				MigrateOutOfOrder:  false,
				DisableColorOutput: false,
				LoggerLevel:        "info",
				MigrationSource:    "migrations",
			}
			bz, err := json.MarshalIndent(&config, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal migration config: %w", err)
			}
			_, err = file.Write(bz)
			if err != nil {
				return fmt.Errorf("failed to write migration config file: %w", err)
			}

			logger := log.AnsiColorLogger{ColorOutput: true, Level: log.LevelInfo}
			logger.Infof("Created migration config file: %s success", filename)
			return nil
		},
	})

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func readConfig(workingDirectoryPtr, configPathPtr *string) (*Config, error) {
	var (
		workingDirectory string
		configPath       string
	)
	if configPathPtr == nil || len(*configPathPtr) == 0 {
		configPath = "migration_config.json"
	} else {
		configPath = *configPathPtr
	}
	if workingDirectoryPtr == nil {
		workingDirectory = ""
	} else {
		workingDirectory = *workingDirectoryPtr
	}

	bz, err := os.ReadFile(path.Join(workingDirectory, configPath))
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(bz, &config)
	if err != nil {
		return nil, err
	}

	if len(workingDirectory) > 0 {
		config.WorkingDirectory = workingDirectory
	}

	if config.SchemaName == "" {
		config.SchemaName = "migration_schema"
	}

	if config.Dialect == "" {
		config.Dialect = "postgres"
	}

	if config.DataSourceName == "" {
		return nil, fmt.Errorf("data_source_name must be set")
	}

	if config.MigrationSource == "" {
		config.MigrationSource = "migrations"
	}

	return &config, nil
}

func convertToMigrateConfig(conf *Config) (*migrate.Config, error) {
	migrateConf := &migrate.Config{
		SchemaName:        conf.SchemaName,
		MigrateOutOfOrder: conf.MigrateOutOfOrder,
		Logger: log.AnsiColorLogger{
			ColorOutput: !conf.DisableColorOutput,
			Level:       log.LevelFromString(conf.LoggerLevel),
		},
		MigrationSource: migrate.DirectoryMigrationSource{Directory: path.Join(conf.WorkingDirectory, conf.MigrationSource)},
	}

	var driverName string
	switch conf.Dialect {
	case "postgres":
		migrateConf.Dialect = migrate.PostgresDialect{}
		driverName = "postgres"
	case "mysql":
		migrateConf.Dialect = migrate.MySQLDialect{}
		driverName = "mysql"
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", conf.Dialect)
	}

	db, err := sql.Open(driverName, conf.DataSourceName)
	if err != nil {
		return nil, err
	}
	migrateConf.DB = db

	return migrateConf, nil
}
