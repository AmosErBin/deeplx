package model

type TranslateReq struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

type TranslateResp struct {
	Code         int         `json:"code"`
	Data         string      `json:"data"`
	Id           int64       `json:"id"`
	Method       string      `json:"method"`
	SourceLang   string      `json:"source_lang"`
	TargetLang   string      `json:"target_lang"`
	Alternatives interface{} `json:"alternatives"`
}
