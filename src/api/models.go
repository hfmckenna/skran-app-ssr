package api

type RecipeItem struct {
	Primary      string
	Sort         string
	Type         string
	Id           uint32
	Title        string
	Components   []Component
	Instructions string
	Time         int16
}

type Component struct {
	Title       string
	Ingredients []Ingredient
}

type Ingredient struct {
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
