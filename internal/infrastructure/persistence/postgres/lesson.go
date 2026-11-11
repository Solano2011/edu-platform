package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"edu-platform/internal/domain/lesson"
)

type LessonRepository struct {
	pool *pgxpool.Pool
}

func NewLessonRepository(pool *pgxpool.Pool) *LessonRepository {
	return &LessonRepository{pool: pool}
}

func (r *LessonRepository) Save(ctx context.Context, l *lesson.Lesson) error {
	query := `
		INSERT INTO lessons (
			id, slug, title, summary, content, subject, 
			content_type, video_url, image_url, reading_time, is_published, 
			published_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			slug = EXCLUDED.slug,
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			content = EXCLUDED.content,
			subject = EXCLUDED.subject,
			content_type = EXCLUDED.content_type,
			video_url = EXCLUDED.video_url,
			image_url = EXCLUDED.image_url,
			reading_time = EXCLUDED.reading_time,
			is_published = EXCLUDED.is_published,
			published_at = EXCLUDED.published_at;
	`
	_, err := r.pool.Exec(ctx, query,
		l.ID(),
		l.Slug(),
		l.Title(),
		l.Summary(),
		l.Body(),
		string(l.Subject()),
		string(l.ContentType()),
		l.VideoURL(),
		l.ImageURL(),
		l.ReadingTime(),
		l.IsPublished(),
		l.PublishedAt(),
	)
	return err
}

func (r *LessonRepository) Update(ctx context.Context, l *lesson.Lesson) error {
	return r.Save(ctx, l)
}

func (r *LessonRepository) FindBySlug(ctx context.Context, slug string) (*lesson.Lesson, error) {
	query := `
		SELECT id, slug, title, summary, content, subject, 
		       content_type, video_url, image_url, reading_time, is_published, 
		       published_at
		FROM lessons
		WHERE slug = $1;
	`
	return r.scanLesson(r.pool.QueryRow(ctx, query, slug))
}

func (r *LessonRepository) FindPublished(ctx context.Context) ([]*lesson.Lesson, error) {
	query := `
		SELECT id, slug, title, summary, content, subject, 
		       content_type, video_url, image_url, reading_time, is_published, 
		       published_at
		FROM lessons
		WHERE is_published = TRUE
		ORDER BY published_at DESC;
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLessons(rows)
}

func (r *LessonRepository) FindPublishedBySubject(ctx context.Context, subject lesson.Subject) ([]*lesson.Lesson, error) {
	query := `
		SELECT id, slug, title, summary, content, subject, 
		       content_type, video_url, image_url, reading_time, is_published, 
		       published_at
		FROM lessons
		WHERE subject = $1 AND is_published = TRUE
		ORDER BY published_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, string(subject))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanLessons(rows)
}

func (r *LessonRepository) Delete(ctx context.Context, slug string) error {
	query := `DELETE FROM lessons WHERE slug = $1;`
	tag, err := r.pool.Exec(ctx, query, slug)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return lesson.ErrLessonNotFound
	}
	return nil
}

func (r *LessonRepository) scanLesson(row pgx.Row) (*lesson.Lesson, error) {
	var (
		id          string
		slug        string
		title       string
		summary     string
		body        string
		subjectStr  string
		cTypeStr    string
		videoURL    string
		imageURL    string
		readingTime int
		isPublished bool
		publishedAt time.Time
	)

	err := row.Scan(
		&id, &slug, &title, &summary, &body, &subjectStr,
		&cTypeStr, &videoURL, &imageURL, &readingTime, &isPublished,
		&publishedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, lesson.ErrLessonNotFound
		}
		return nil, err
	}

	return lesson.Restore(
		id, slug, title, summary, body,
		lesson.Subject(subjectStr),
		lesson.ContentType(cTypeStr),
		videoURL, imageURL, readingTime,
		publishedAt, isPublished,
	), nil
}

func (r *LessonRepository) scanLessons(rows pgx.Rows) ([]*lesson.Lesson, error) {
	var lessons []*lesson.Lesson
	for rows.Next() {
		var (
			id          string
			slug        string
			title       string
			summary     string
			body        string
			subjectStr  string
			cTypeStr    string
			videoURL    string
			imageURL    string
			readingTime int
			isPublished bool
			publishedAt time.Time
		)

		if err := rows.Scan(
			&id, &slug, &title, &summary, &body, &subjectStr,
			&cTypeStr, &videoURL, &imageURL, &readingTime, &isPublished,
			&publishedAt,
		); err != nil {
			return nil, err
		}

		lessons = append(lessons, lesson.Restore(
			id, slug, title, summary, body,
			lesson.Subject(subjectStr),
			lesson.ContentType(cTypeStr),
			videoURL, imageURL, readingTime,
			publishedAt, isPublished,
		))
	}

	return lessons, rows.Err()
}
