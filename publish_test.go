package publish_test

import (
	"fmt"

	"github.com/moisespsena-go/aorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ecletus/l10n"
	"github.com/ecletus/publish"
	"github.com/ecletus/core/test/utils"
)

var pb *publish.Publish
var pbdraft *aorm.DB
var pbprod *aorm.DB
var db *aorm.DB

func init() {
	db = utils.TestDB()
	l10n.RegisterCallbacks(db)

	pb = publish.New(db)
	pbdraft = pb.DraftDB()
	pbprod = pb.ProductionDB()

	for _, table := range []string{"product_categories", "product_categories_draft", "product_languages", "product_languages_draft", "author_books", "author_books_draft"} {
		pbprod.Exec(fmt.Sprintf("drop table %v", table))
	}

	for _, value := range []interface{}{&Product{}, &Color{}, &Category{}, &Language{}, &Book{}, &Publisher{}, &Comment{}, &Author{}} {
		pbprod.DropTable(value)
		pbdraft.DropTable(value)

		pbprod.AutoMigrate(value)
		pb.AutoMigrate(value)
	}
}

type Product struct {
	aorm.Model
	Name       string
	Quantity   uint
	Color      Color
	ColorId    int
	Categories []Category `gorm:"many2many:product_categories"`
	Languages  []Language `gorm:"many2many:product_languages"`
	publish.Status
}

type Color struct {
	aorm.Model
	Name string
}

type Language struct {
	aorm.Model
	Name string
}

type Category struct {
	aorm.Model
	Name string
	publish.Status
}
