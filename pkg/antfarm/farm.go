package antfarm

type Room struct{
	Name  string
	X  int
	Y  int
	IsStart  bool
	IsEnd    bool
}

type Farm struct{
	Ants int
	Start *Room
	End *Room
	Rooms map[string]*Room 
	Links map[string][]string
	RawInput  string
}

//new farm initializes a new farm struct
func NewFarm() *Farm{
	return &Farm{
		Rooms: make(map[string] *Room),
		Links: make(map[string][]string),
	}
}

//Addlink adds an undirected link between 2 rooms
func (f *Farm) AddLink(u, v string){
	f.Links[u] = append(f.Links[u], v)
	f.Links[v] = append(f.Links[v], u)
}

//path represents a route from ##start to ##end
type Path []string

type Movement struct{
	AntID int
	RoomName string
}