package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestResolvePageIDRejectsDuplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM page WHERE name=? AND status=1 AND deleted_at IS NULL ORDER BY id LIMIT 2")).WithArgs("资讯动态").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
	if _, err := ResolvePageID(context.Background(), db, "资讯动态"); err != ErrPageNotUnique {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchArticleColumnsUsesPublishableTypeAndReturnsPageName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("(?s)SELECT DISTINCT c.id.*a.status=1.*a.audit_status=2.*a.deleted_at IS NULL.*a.type IN \\(1,2\\)").
		WithArgs(int64(1001)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "page_id", "page_name", "parent_id", "description", "sort"}).
			AddRow(21, "中心动态", "center", 1, "资讯动态", 0, "", 10).
			AddRow(41, "招聘信息", "recruitment", 4, "关于我们", 0, "", 20))
	columns, err := New(db, 1).FetchArticleColumns(context.Background(), 1001)
	if err != nil {
		t.Fatal(err)
	}
	if len(columns) != 2 || columns[0].PageName != "资讯动态" || columns[1].PageName != "关于我们" {
		t.Fatalf("unexpected columns: %+v", columns)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFetchColumnsByIDsRejectsMissingColumn(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("(?s)FROM `column` c.*c.id IN \\(\\?,\\?\\).*").
		WithArgs(int64(21), int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "page_id", "page_name", "parent_id", "description", "sort"}).
			AddRow(21, "中心动态", "center", 1, "资讯动态", 0, "", 10))
	_, err = New(db, 1).FetchColumnsByIDs(context.Background(), []int64{99, 21, 21})
	if !errors.Is(err, ErrColumnNotFound) {
		t.Fatalf("error = %v, want ErrColumnNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFetchByColumnUsesPublishedFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectQuery("SELECT id,name,code,page_id,parent_id").WithArgs(int64(7), "中心动态").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "page_id", "parent_id", "description", "sort"}).AddRow(21, "中心动态", "center", 7, 0, "", 10))
	mock.ExpectQuery("(?s)FROM article_column_publish acp.*a.status=1.*a.audit_status=2.*a.deleted_at IS NULL.*a.type IN \\(1,2\\).*ORDER BY acp.is_top DESC").WithArgs(int64(21), 10).WillReturnRows(sqlmock.NewRows([]string{"id", "type", "title", "summary", "content", "cover", "author", "source", "is_bold", "is_top", "default_color", "url", "publish_time"}).AddRow(1001, 1, "标题", "摘要", "正文", "", "", "中心", 0, 1, "", "", time.Now()))
	_, items, err := New(db, 7).FetchByColumnName(context.Background(), "中心动态", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ColumnID != 21 {
		t.Fatalf("unexpected items: %+v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
