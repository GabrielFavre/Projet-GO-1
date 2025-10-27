package main

import (
	"html/template"
)

func init() {
	funcMap := template.FuncMap{
		"iterate": func(count int) []int {
			var items []int
			for i := 0; i < count; i++ {
				items = append(items, i)
			}
			return items
		},
		"add": func(a, b int) int {
			return a + b
		},
		"eq": func(a, b int) bool {
			return a == b
		},
	}

	templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
}
