package antfarm

type Room struct{
	Name  string
	X, Y  int
	IsStart  bool
	IsEnd    bool
}

type Farm struct{
	Ants int
	Start *Room
	End *Room
	Rooms map[string]*Room 
	Links map[string][]string
}