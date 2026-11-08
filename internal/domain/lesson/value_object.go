package lesson

type ContentType string

const (
	TypeArticle ContentType = "article"
	TypeVideo   ContentType = "video"
	TypeMixed   ContentType = "mixed"
)

func (ct ContentType) IsValid() bool {
	switch ct {
	case TypeArticle, TypeVideo, TypeMixed:
		return true
	default:
		return false
	}
}

// Subject — Value Object учебного направления (предмета)
type Subject string

const (
	SubjectHistory Subject = "history"
	SubjectSocial  Subject = "social"
)

func (s Subject) IsValid() bool {
	switch s {
	case SubjectHistory, SubjectSocial:
		return true
	default:
		return false
	}
}

func (s Subject) DisplayName() string {
	switch s {
	case SubjectHistory:
		return "История"
	case SubjectSocial:
		return "Обществознание"
	default:
		return "Неизвестное направление"
	}
}
