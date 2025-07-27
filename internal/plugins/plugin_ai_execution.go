package plugins

type AIExecToolParams struct {
	Code string `json:"code"`
}

type AIUserToolParams struct {
	Args []string `json:"args"`
}
