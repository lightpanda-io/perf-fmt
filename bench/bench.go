package bench

type InItem struct {
	Duration  int `json:"duration"`
	AllocSize int `json:"alloc_size"`
	AllocNb   int `json:"alloc_nb"`
	ReallocNb int `json:"realloc_nb"`
	FreeNb    int `json:"free_nb"`
}

type InResult struct {
	Name  string `json:"name"`
	Bench InItem `json:"bench"`
}

type OutItem struct {
	Duration  int `json:"duration"`
	AllocSize int `json:"alloc_size"`
	AllocNb   int `json:"alloc_nb"`
	ReallocNb int `json:"realloc_nb"`
	FreeNb    int `json:"free"`
}
