package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"
)

type IDemodataProvider interface {
	Up(ctx *DBContext)
	Down(ctx *DBContext)
}

type AppDataProvider struct {
	folder string
}

type DemodataProvider struct{}

func NewDemodataProvider() *DemodataProvider {
	return &DemodataProvider{}
}

func NewAppDataProvider(folder string) *AppDataProvider {
	return &AppDataProvider{folder}
}

func (d AppDataProvider) Up(ctx *DBContext) {
	must(InitializeDemodata[User](ctx, path.Join(d.folder, "users.json")))
	must(InitializeDemodata[Project](ctx, path.Join(d.folder, "projects.json")))
	must(InitializeDemodata[File](ctx, path.Join(d.folder, "files.json")))
	must(InitializeDemodata[Comment](ctx, path.Join(d.folder, "comments.json")))
	must(d.items(ctx))
}

func (d AppDataProvider) Down(ctx *DBContext) {
	must(ctx.DB.Delete(&Item{}, "1 = 1").Error)
	must(ctx.DB.Delete(&ItemUser{}, "1 = 1").Error)
	must(ctx.DB.Delete(&Project{}, "1 = 1").Error)
	must(ctx.DB.Delete(&User{}, "1 = 1").Error)
	must(ctx.DB.Delete(&File{}, "1 = 1").Error)
	must(ctx.DB.Delete(&Comment{}, "1 = 1").Error)
}

func (d AppDataProvider) items(ctx *DBContext) error {
	var items []Item
	err := ParseJSON(path.Join(d.folder, "items.json"), &items)
	if err != nil {
		return err
	}

	shiftItemDates(items)

	indexMap := make(map[string]int)
	for i := range items {
		item := &items[i]
		key := fmt.Sprintf("%d:%d", item.ProjectID, item.ParentID)
		item.Index = indexMap[key]
		indexMap[key]++
	}

	return ctx.DB.Create(&items).Error
}

// demodata dates are fixed around Sep-Dec 2024; shift them by a whole number
// of weeks so that the busiest demo week (Nov 11, 2024) lands on the current
// week and weekdays stay aligned
func shiftItemDates(items []Item) {
	anchor := time.Date(2024, 11, 11, 0, 0, 0, 0, time.UTC)
	days := 7 * int(time.Since(anchor).Hours()/(24*7))
	if days <= 0 {
		return
	}

	for i := range items {
		item := &items[i]
		dates := []**time.Time{
			&item.StartDate, &item.EndDate,
			&item.CreationDate, &item.EditedDate, &item.CompletionDate,
		}
		for _, d := range dates {
			if *d != nil {
				shifted := (*d).AddDate(0, 0, days)
				*d = &shifted
			}
		}
	}
}

func (d DemodataProvider) Restore(ctx *DBContext, providers ...IDemodataProvider) {
	for i := range providers {
		providers[i].Down(ctx)
		providers[i].Up(ctx)
	}
}

func (d DemodataProvider) Down(ctx *DBContext, providers ...IDemodataProvider) {
	for i := range providers {
		providers[i].Down(ctx)
	}
}

func (d DemodataProvider) Up(ctx *DBContext, providers ...IDemodataProvider) {
	for i := range providers {
		providers[i].Up(ctx)
	}
}

func InitializeDemodata[T any](ctx *DBContext, path string) error {
	data := make([]T, 0)
	err := ParseJSON(path, &data)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}

	err = ctx.DB.Create(&data).Error

	if Debug {
		if err == nil {
			fmt.Printf("--- Initialized %d rows of %T\n", len(data), data)
		}
	}

	return err
}

func ParseJSON(path string, dest any) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, &dest)
}
