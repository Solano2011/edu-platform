package lesson

import (
	"strings"
	"time"
)

type Lesson struct {
	id          string
	slug        string
	title       string
	summary     string
	body        string
	subject     Subject
	contentType ContentType
	videoURL    string
	imageURL    string // Путь к загруженной картинке (например, /static/uploads/map.png)
	readingTime int
	publishedAt time.Time
	isPublished bool
}

func NewLesson(
	id, slug, title, summary, body string,
	subject Subject,
	contentType ContentType,
	videoURL, imageURL string, // Добавили параметр imageURL
	readingTime int,
) (*Lesson, error) {
	if strings.TrimSpace(title) == "" {
		return nil, ErrEmptyTitle
	}
	if !subject.IsValid() {
		return nil, ErrInvalidSubject
	}
	if !contentType.IsValid() {
		return nil, ErrInvalidType
	}

	return &Lesson{
		id:          id,
		slug:        slug,
		title:       title,
		summary:     summary,
		body:        body,
		subject:     subject,
		contentType: contentType,
		videoURL:    videoURL,
		imageURL:    imageURL,
		readingTime: readingTime,
		publishedAt: time.Now(),
		isPublished: true,
	}, nil
}

func (l *Lesson) Update(
	title, summary, body string,
	subject Subject,
	contentType ContentType,
	videoURL, imageURL string,
	readingTime int,
) error {
	if strings.TrimSpace(title) == "" {
		return ErrEmptyTitle
	}
	if !subject.IsValid() {
		return ErrInvalidSubject
	}
	if !contentType.IsValid() {
		return ErrInvalidType
	}

	l.title = title
	l.summary = summary
	l.body = body
	l.subject = subject
	l.contentType = contentType
	l.videoURL = videoURL
	if imageURL != "" {
		l.imageURL = imageURL
	}
	l.readingTime = readingTime

	return nil
}

func (l *Lesson) ID() string               { return l.id }
func (l *Lesson) Slug() string             { return l.slug }
func (l *Lesson) Title() string            { return l.title }
func (l *Lesson) Summary() string          { return l.summary }
func (l *Lesson) Body() string             { return l.body }
func (l *Lesson) Subject() Subject         { return l.subject }
func (l *Lesson) ContentType() ContentType { return l.contentType }
func (l *Lesson) VideoURL() string         { return l.videoURL }
func (l *Lesson) ImageURL() string         { return l.imageURL } // Геттер для картинки
func (l *Lesson) ReadingTime() int         { return l.readingTime }
func (l *Lesson) PublishedAt() time.Time   { return l.publishedAt }
func (l *Lesson) IsPublished() bool        { return l.isPublished }
