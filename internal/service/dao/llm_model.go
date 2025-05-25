package dao

type llmModelDao struct {
	*daoService
}

var LlmModelDao = llmModelDao{
	daoService: Dao,
}
