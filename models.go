package main

type RecipeItem struct {
	Primary      string
	Sort         string
	Id           uint32
	Title        string
	Component    []Component
	Instructions string
	Type         string
}

type Component struct {
	Title       string
	Value       uint16
	Measurement string
}

type SearchItem struct {
	Primary     string
	Sort        string
	Id          string
	Title       string
	Ingredients []string
	Type        string
}
