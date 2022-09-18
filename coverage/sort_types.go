package main

type orderNameAsc []User

func (o orderNameAsc) Len() int {
	return len(o)
}

// Less определяет по какому принципу перемешать элементы влево(в начало) слайса. Здесь влево пойдет меньший элемент.
func (o orderNameAsc) Less(i, j int) bool {
	return o[i].Name < o[j].Name
}

func (o orderNameAsc) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

type orderNameDesc []User

func (o orderNameDesc) Len() int {
	return len(o)
}

// Less определяет по какому принципу перемешать элементы влево(в начало) слайса. Здесь влево пойдет больший элемент.
func (o orderNameDesc) Less(i, j int) bool {
	return o[i].Name > o[j].Name
}

func (o orderNameDesc) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

type orderIdAsc []User

func (o orderIdAsc) Len() int {
	return len(o)
}

// Less определяет по какому принципу перемешать элементы влево(в начало) слайса. Здесь влево пойдет меньший элемент.
func (o orderIdAsc) Less(i, j int) bool {
	return o[i].ID < o[j].ID
}

func (o orderIdAsc) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

type orderIdDesc []User

func (o orderIdDesc) Len() int {
	return len(o)
}

// Less определяет по какому принципу перемешать элементы влево(в начало) слайса. Здесь влево пойдет больший элемент.
func (o orderIdDesc) Less(i, j int) bool {
	return o[i].ID > o[j].ID
}

func (o orderIdDesc) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

type orderAgeAsc []User

func (o orderAgeAsc) Len() int {
	return len(o)
}

// Less определяет по какому принципу перемешать элементы влево(в начало) слайса. Здесь влево пойдет меньший элемент.
func (o orderAgeAsc) Less(i, j int) bool {
	return o[i].Age < o[j].Age
}

func (o orderAgeAsc) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

type orderAgeDesc []User

func (o orderAgeDesc) Len() int {
	return len(o)
}

// Less определяет по какому принципу перемешать элементы влево(в начало) слайса. Здесь влево пойдет больший элемент.
func (o orderAgeDesc) Less(i, j int) bool {
	return o[i].Age > o[j].Age
}

func (o orderAgeDesc) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}
