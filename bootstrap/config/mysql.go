package config

type MySQLConfig struct {
	DSN                           string
	SkipInitializeWithVersion     bool
	DefaultStringSize             uint
	DefaultDatetimePrecision      *int
	DisableWithReturning          bool
	DisableDatetimePrecision      bool
	DontSupportRenameIndex        bool
	DontSupportRenameColumn       bool
	DontSupportForShareClause     bool
	DontSupportNullAsDefaultValue bool
	DontSupportRenameColumnUnique bool
	// As of MySQL 8.0.19, ALTER TABLE permits more general (and SQL standard) syntax
	// for dropping and altering existing constraints of any type.
	// see https://dev.mysql.com/doc/refman/8.0/en/alter-table.html
	DontSupportDropConstraint bool
	DebugMode                 bool
	LogLevel                  string
}
