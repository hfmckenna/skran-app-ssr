package models

type RecipeItem struct {
	Primary      string      `json:"Primary"`
	Sort         string      `json:"Sort"`
	Type         string      `json:"Type"`
	Id           string      `json:"Id"`
	Title        string      `json:"Title"`
	Components   []Component `json:"Components"`
	Instructions string      `json:"Instructions"`
	Time         int16       `json:"Time"`
}

type Component struct {
	Title       string       `json:"Title"`
	Ingredients []Ingredient `json:"Ingredients"`
}

type Ingredient struct {
	Title       string `json:"Title"`
	Value       uint16 `json:"Value"`
	Measurement string `json:"Measurement"`
}

type SearchItem struct {
	Primary string   `json:"Primary"`
	Sort    string   `json:"Sort"`
	Title   string   `json:"Title"`
	Recipes []string `json:"Recipes"`
	Type    string   `json:"Type"`
}
