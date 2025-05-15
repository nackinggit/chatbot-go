package config

type MySQLConfig struct {
	RegisterName string `json:"registerName" yaml:"registerName" mapstructure:"registerName"`
	DSN          string `json:"dsn" yaml:"dsn" mapstructure:"dsn"`
	DebugMode    bool   `json:"debugMode" yaml:"debugMode" mapstructure:"debugMode"`
	LogLevel     string `json:"logLevel" yaml:"logLevel" mapstructure:"logLevel"`

	SkipInitializeWithVersion     bool `json:"skipInitializeWithVersion" yaml:"skipInitializeWithVersion" mapstructure:"skipInitializeWithVersion"`
	DefaultStringSize             uint `json:"defaultStringSize" yaml:"defaultStringSize" mapstructure:"defaultStringSize"`
	DefaultDatetimePrecision      *int `json:"defaultDatetimePrecision" yaml:"defaultDatetimePrecision" mapstructure:"defaultDatetimePrecision"`
	DisableWithReturning          bool `json:"disableWithReturning" yaml:"disableWithReturning" mapstructure:"disableWithReturning"`
	DisableDatetimePrecision      bool `json:"disableDatetimePrecision" yaml:"disableDatetimePrecision" mapstructure:"disableDatetimePrecision"`
	DontSupportRenameIndex        bool `json:"dontSupportRenameIndex" yaml:"dontSupportRenameIndex" mapstructure:"dontSupportRenameIndex"`
	DontSupportRenameColumn       bool `json:"dontSupportRenameColumn" yaml:"dontSupportRenameColumn" mapstructure:"dontSupportRenameColumn"`
	DontSupportForShareClause     bool `json:"dontSupportForShareClause" yaml:"dontSupportForShareClause" mapstructure:"dontSupportForShareClause"`
	DontSupportNullAsDefaultValue bool `json:"dontSupportNullAsDefaultValue" yaml:"dontSupportNullAsDefaultValue" mapstructure:"dontSupportNullAsDefaultValue"`
	DontSupportRenameColumnUnique bool `json:"dontSupportRenameColumnUnique" yaml:"dontSupportRenameColumnUnique" mapstructure:"dontSupportRenameColumnUnique"`
	// As of MySQL 8.0.19, ALTER TABLE permits more general (and SQL standard) syntax
	// for dropping and altering existing constraints of any type.
	// see https://dev.mysql.com/doc/refman/8.0/en/alter-table.html
	DontSupportDropConstraint bool `json:"dontSupportDropConstraint" yaml:"dontSupportDropConstraint" mapstructure:"dontSupportDropConstraint"`
}
