package main

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

type Dinosaur struct {
	Name string
	Era  string
}

func (d *Dinosaur) String() string {
	return fmt.Sprintf("%s (%s)", d.Name, d.Era)
}

func (d *Dinosaur) Roar() string {
	return fmt.Sprintf("%s roars!", d.Name)
}

func checkDinosaur(L *lua.LState) *Dinosaur {
	ud := L.CheckUserData(1)
	if o, ok := ud.Value.(*Dinosaur); ok {
		return o
	}
	L.ArgError(1, "expected a dinosaur")
	return nil
}

func newDinosaur(L *lua.LState) int {
	name := L.CheckString(1)
	era := L.CheckString(2)
	dino := &Dinosaur{Name: name, Era: era}
	ud := L.NewUserData()
	ud.Value = dino
	L.SetMetatable(ud, L.GetTypeMetatable("dinosaur"))
	L.Push(ud)
	return 1
}

func registerDinosaurType(L *lua.LState) {
	mt := L.NewTypeMetatable("dinosaur")
	L.SetGlobal("dinosaur", mt)

	L.SetField(mt, "new", L.NewFunction(newDinosaur))
	L.SetField(mt, "__tostring", L.NewFunction(func(L *lua.LState) int {
		dino := checkDinosaur(L)
		L.Push(lua.LString(dino.String()))
		return 1
	}))
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"roar": func(L *lua.LState) int {
			dino := checkDinosaur(L)
			L.Push(lua.LString(dino.Roar()))
			return 1
		},
	}))
}

var dinosaurs = []Dinosaur{
	{Name: "Tyrannosaurus", Era: "Late Cretaceous"},
	{Name: "Velociraptor", Era: "Late Cretaceous"},
	{Name: "Triceratops", Era: "Late Cretaceous"},
}

func iterateLuaTableInGo(L *lua.LState) int {
	iter := L.Get(1).(*lua.LTable)
	for k, v := iter.Next(lua.LNil); k != lua.LNil; k, v = iter.Next(k) {
		fmt.Println("key", k, "value", v)
	}
	return 0
}

type Tree struct {
	Name     string
	Age      int
	Children []*Tree
}

func (t *Tree) AddChild(child *Tree) {
	t.Children = append(t.Children, child)
}

func createLuaStateWithLuar() *lua.LState {
	L := lua.NewState()

	// Register Dinosaur type
	L.SetGlobal("Dinosaur", luar.New(L, Dinosaur{}))

	// Register Dinosaur methods
	L.SetGlobal("newDinosaur", luar.New(L, func(name, era string) *Dinosaur {
		return &Dinosaur{Name: name, Era: era}
	}))

	// Register Tree type
	L.SetGlobal("Tree", luar.NewType(L, Tree{}))

	return L
}

func luaTableToGoTree(L *lua.LState, table *lua.LTable) *Tree {
	tree := &Tree{}

	table.ForEach(func(key, value lua.LValue) {
		switch key.String() {
		case "Name":
			tree.Name = value.String()
		case "Age":
			tree.Age = int(value.(lua.LNumber))
		case "Children":
			if childrenTable, ok := value.(*lua.LTable); ok {
				childrenTable.ForEach(func(_, childValue lua.LValue) {
					if childTable, ok := childValue.(*lua.LTable); ok {
						tree.Children = append(tree.Children, luaTableToGoTree(L, childTable))
					}
				})
			}
		}
	})

	return tree
}

func main() {
	L := lua.NewState()
	defer L.Close()

	tbl := L.NewTable()

	L.SetGlobal("tbl", tbl)
	tbl.RawSetString("name", lua.LString("gopher"))
	tbl.RawSetInt(1, lua.LString("gopherInt"))
	tbl.RawSet(lua.LNumber(2), lua.LString("gopherNumber"))

	luaScript := `
	    print (tbl)
		print(tbl.name)
		print(tbl[1])
		print(tbl[2])
	`

	if err := L.DoString(luaScript); err != nil {
		panic(err)
	}

	iterateFn := L.NewFunction(iterateLuaTableInGo)
	L.SetGlobal("iterateFn", iterateFn)

	L.CallByParam(lua.P{
		Fn:      iterateFn,
		NRet:    0,
		Protect: true,
	}, tbl)

	luaScript = `
		iterateFn(tbl)
	`

	dinosaurTbl := L.NewTable()
	for _, dino := range dinosaurs {
		dinosaurTbl.RawSetString(dino.Name, lua.LString(dino.String()))
	}

	L.SetGlobal("dinosaurs", dinosaurTbl)
	registerDinosaurType(L)

	luaScript = `
		iterateFn(dinosaurs)
		print("---")
		print(dinosaur.new)
		local dino = dinosaur.new("Tyrannosaurus", "Late Cretaceous")
		print("---")
		print(dino)
		local mt = getmetatable(dino)
		print(mt)
		-- iterate over mt
		for k, v in pairs(mt) do
			print(k, v)
		end
		print("---")
		print(dino)
		print(dino.name)
		print("---")
		print(dino:roar())
	`

	if err := L.DoString(luaScript); err != nil {
		panic(err)
	}

	// Create a new Lua state with gopher-luar
	L2 := createLuaStateWithLuar()
	defer L2.Close()

	luarScript := `
		local dino = newDinosaur("Stegosaurus", "Late Jurassic")
		print(dino:String())
		print(dino:Roar())

		local oak = Tree()
		oak.Name = "Oak"
		oak.Age = 100

		local sapling1 = Tree()
		sapling1.Name = "Sapling 1"
		sapling1.Age = 5

		local sapling2 = Tree()
		sapling2.Name = "Sapling 2"
		sapling2.Age = 3

		local seedling = Tree()
		seedling.Name = "Seedling"
		seedling.Age = 1

		oak:AddChild(sapling1)
		oak:AddChild(sapling2)
		sapling2:AddChild(seedling)

		return oak
	`

	if err := L2.DoString(luarScript); err != nil {
		panic(err)
	}

	// Print the type of L2.Get(-1)
	fmt.Printf("Type of L2.Get(-1): %T\n", L2.Get(-1))

	// // Get the returned Tree instance
	luaTree := L2.Get(-1).(*lua.LUserData)
	goTree := luaTree.Value.(*Tree)

	fmt.Printf("Converted Tree: %+v\n", goTree)
	fmt.Printf("First Child: %+v\n", goTree.Children[0])
	fmt.Printf("Second Child: %+v\n", goTree.Children[1])
	fmt.Printf("Grandchild: %+v\n", goTree.Children[1].Children[0])
}
