package types

// BindingType 绑定类型枚举
type BindingType string

const (
	BindingFunc   BindingType = "func"   // 函数
	BindingStruct BindingType = "struct" // 结构体
	BindingConst  BindingType = "const"  // 常量
	BindingVar    BindingType = "var"    // 变量
)

// Binding 表示文档与代码的绑定关系
type Binding struct {
	// 文档位置
	DocFile    string `json:"doc_file"`
	DocLine    int    `json:"doc_line"`
	DocContent string `json:"doc_content"`

	// 代码位置
	CodeFile   string      `json:"code_file"`
	CodeSymbol string      `json:"code_symbol"`
	CodeType   BindingType `json:"code_type"`

	// 解析后的代码内容
	CodeContent string `json:"code_content,omitempty"`
}

// Location 位置信息
type Location struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column,omitempty"`
	Symbol string `json:"symbol,omitempty"`
}
