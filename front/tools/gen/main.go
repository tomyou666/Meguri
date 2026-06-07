package main

import (
	"os"

	"github.com/libtnb/sqlite"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// stripIntDefault removes gorm default tags on INTEGER 0/1 columns so SQLite
// batch INSERT emits literal 0/1 instead of DEFAULT (unsupported in VALUES).
func stripIntDefault(column string) gen.ModelOpt {
	return gen.FieldGORMTag(column, func(tag field.GormTag) field.GormTag {
		tag.Remove("default")
		return tag
	})
}

func main() {
	dbPath := os.Getenv("DB")
	if dbPath == "" {
		dbPath = "data/scraperbot.db"
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	g := gen.NewGenerator(gen.Config{
		OutPath:       "./internal/query",
		Mode:          gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable: true,
	})
	g.UseDB(db)
	g.ApplyBasic(
		g.GenerateModel("app_config"),
		g.GenerateModel("workspaces"),
		g.GenerateModel("graph_nodes",
			stripIntDefault("user_positioned"),
			stripIntDefault("crawl_exclude"),
		),
		g.GenerateModel("graph_edges"),
		g.GenerateModel("domain_settings"),
		g.GenerateModel("crawl_runs"),
		g.GenerateModel("node_results"),
		g.GenerateModel("graph_ui_state"),
	)
	g.Execute()
}
