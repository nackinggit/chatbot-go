package xmysql

import (
	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/util"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbs map[string]*gorm.DB = map[string]*gorm.DB{}

func GetDB(name string) *gorm.DB {
	return dbs[name]
}

func Init(cfgs []*config.MySQLConfig) {
	convertCfg := func(cfg *config.MySQLConfig) (mysql.Config, bool) {
		return mysql.Config{
			DSN:                           cfg.DSN,
			SkipInitializeWithVersion:     cfg.SkipInitializeWithVersion,
			DefaultStringSize:             cfg.DefaultStringSize,
			DefaultDatetimePrecision:      cfg.DefaultDatetimePrecision,
			DisableWithReturning:          cfg.DisableWithReturning,
			DisableDatetimePrecision:      cfg.DisableDatetimePrecision,
			DontSupportRenameIndex:        cfg.DontSupportRenameIndex,
			DontSupportRenameColumn:       cfg.DontSupportRenameColumn,
			DontSupportForShareClause:     cfg.DontSupportForShareClause,
			DontSupportNullAsDefaultValue: cfg.DontSupportNullAsDefaultValue,
			DontSupportRenameColumnUnique: cfg.DontSupportRenameColumnUnique,
			// As of MySQL 8.0.19, ALTER TABLE permits more general (and SQL standard) syntax
			// for dropping and altering existing constraints of any type.
			// see https://dev.mysql.com/doc/refman/8.0/en/alter-table.html
			DontSupportDropConstraint: cfg.DontSupportDropConstraint,
		}, cfg.DebugMode
	}
	xlog.Infof("init mysql: %v", util.JsonString(cfgs))
	for _, cfg := range cfgs {
		mcfg, debug := convertCfg(cfg)
		level := logger.Info
		switch cfg.LogLevel {
		case "info":
			level = logger.Info
		case "warn":
			level = logger.Warn
		case "error":
			level = logger.Error
		case "silent":
			level = logger.Silent
		}
		dbs[cfg.RegisterName] = NewMysql(mcfg, debug, logger.Default.LogMode(level))
	}
}

func NewMysql(cfg mysql.Config, debugMode bool, logger logger.Interface) *gorm.DB {
	db, err := gorm.Open(mysql.New(cfg), &gorm.Config{Logger: logger})
	if err != nil {
		panic(err)
	}
	if debugMode {
		db = db.Debug()
	}
	return db
}
