package main

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

func main() {
	L := lua.NewState()
	defer L.Close()

	luaScript := `
	function count() 
	for i = 1, 3 do
		print(i)
		coroutine.yield()
	end
	print("Done!")
	end	
	`

	if err := L.DoString(luaScript); err != nil {
		panic(err)
	}

	countFunc := L.GetGlobal("count").(*lua.LFunction)
	co, _ := L.NewThread()

	for {
		st, err, values := L.Resume(co, countFunc)
		if err != nil {
			panic(err)
		}
		for i, lv := range values {
			fmt.Println("value", i, lv)
		}
		fmt.Println("status", st)
		if st == lua.ResumeOK {
			break
		}
	}

	luaScript = `
	function countTo(a) 
	   return function() 
	      for i = 1, a do
			print(i)
			coroutine.yield(i)
		  end
		end
	end
	`

	if err := L.DoString(luaScript); err != nil {
		panic(err)
	}

	countToFunc := L.GetGlobal("countTo").(*lua.LFunction)

	co, _ = L.NewThread()

	if err := L.CallByParam(lua.P{
		Fn:      countToFunc,
		NRet:    1,
		Protect: true,
	}, lua.LNumber(10)); err != nil {
		panic(err)
	}
	countTo10 := L.Get(-1).(*lua.LFunction)
	L.Pop(1)

	for {
		st, err, values := L.Resume(co, countTo10)
		if err != nil {
			panic(err)
		}
		for i, lv := range values {
			fmt.Println("value", i, lv)
		}
		fmt.Println("status", st, L.Status(co))

		if st == lua.ResumeOK {
			break
		}
	}

}
