package service

type ServiceConfig struct {
	Teacher *TeacherConfig `json:"teacher" yaml:"teacher" mapstructure:"teacher"` // 解题助手配置
}

type TeacherConfig struct {
	QuestionAnalyse string   `json:"questionAnalyse" yaml:"questionAnalyse" mapstructure:"questionAnalyse"` // 解析题目配置
	AnswerModels    []string `json:"answerModels" yaml:"answerModels" mapstructure:"answerModels"`          // 解析答案模型配置
	JudgeModel      string   `json:"judgeModel" yaml:"judgeModel" mapstructure:"judgeModel"`                // 判断模型配置
}
