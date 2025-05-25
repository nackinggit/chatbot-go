package dao

import (
	"context"

	"com.imilair/chatbot/internal/model/dbmodel"
)

type llmModelDao struct {
	*daoService
}

var LlmModelDao = llmModelDao{
	daoService: Dao,
}

func (d *llmModelDao) QueryChathistoryBySids(ctx context.Context, sids []string) ([]*dbmodel.LlmChatHistory, error) {
	var chhs = []*dbmodel.LlmChatHistory{}
	err := d.GetDbTableByModel(ctx, &dbmodel.LlmChatHistory{}).
		Where("sid IN (?)", sids).
		Order("id").
		Find(&chhs).Error
	return chhs, err
}
