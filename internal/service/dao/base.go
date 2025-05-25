package dao

import (
	"context"
	"fmt"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/pkg/xmysql"

	"gorm.io/gorm"
)

type daoService struct {
	chatbotdb *gorm.DB
}

var Dao *daoService

func (t *daoService) Name() string {
	return "dao"
}

func (t *daoService) InitAndStart() error {
	cfg := service.Config.Dao
	if cfg == nil {
		return fmt.Errorf("init service `%s` failed, config is nil", t.Name())
	}
	xlog.Infof("init service `%s`", t.Name())
	t.chatbotdb = xmysql.GetDB(cfg.DbName)
	if t.chatbotdb == nil {
		return fmt.Errorf("init service `%s` failed, %s not inited", t.Name(), cfg.DbName)
	}
	Dao = t
	return nil
}

func (t *daoService) Stop() {}

func init() {
	service.Register(&daoService{})
}

type (
	// 用来承载事务的上下文
	contextTxKey struct{}
)

func (dao *daoService) ExecSqlTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return dao.chatbotdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, contextTxKey{}, tx)
		return fn(ctx)
	})
}

func (d *daoService) TxDB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if ok {
		return tx
	}
	return d.chatbotdb
}

func (dao *daoService) GetDbTableByModel(ctx context.Context, dbmodel interface{}) *gorm.DB {
	return dao.TxDB(ctx).Model(dbmodel)
}

func (dao *daoService) GetDbTableByName(ctx context.Context, modelname string) *gorm.DB {
	return dao.TxDB(ctx).Table(modelname)
}

func QueryById[T any](ctx context.Context, t T, id any) (*T, error) {
	err := Dao.GetDbTableByModel(ctx, &t).Where("id = ?", id).First(&t).Error
	return &t, err
}
