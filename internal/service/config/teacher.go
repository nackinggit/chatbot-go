package config

type TeacherConfig struct {
	QuestionAnalyse *BotConfig   `json:"questionAnalyse" yaml:"questionAnalyse" mapstructure:"questionAnalyse"` // 解析题目配置
	AnswerModels    []*BotConfig `json:"answerModels" yaml:"answerModels" mapstructure:"answerModels"`          // 解析答案模型配置
	JudgeModel      *BotConfig   `json:"judgeModel" yaml:"judgeModel" mapstructure:"judgeModel"`                // 判断模型配置
}
