package main


import(
	"fmt"
	"os"
	
	"lem-in/pkg/parser"
	"lem-in/pkg/solver"
)

func main(){
	if len(os.Args)!=2{
		fmt.Println("ERROR: invalid data format, usage: go run . <filename>")
		os.Exit(1)
	}
	filePath:= os.Args[1]
	farm, err:= parser.ParseFile(filePath)
	if err != nil{
		fmt.Println(err.Error())
		os.Exit(1)
	}

	//find the paths
	paths, err:= solver.FindOptimalPaths(farm)
	if err!=nil{
		fmt.Println(err.Error())
		os.Exit(1)
	}

	moves:= solver.DispatchAndSimulate(paths, farm.Ants)

	//print raw file content
	fmt.Println(farm.RawInput)
	fmt.Println()

	//print ant turn-by-turn movements
	for _, line:= range moves{
		fmt.Println(line)
	}
}