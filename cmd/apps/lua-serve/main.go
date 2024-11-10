package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-go-golems/clay/pkg/watcher"
	lua2 "github.com/go-go-golems/glazed/pkg/lua"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	lua "github.com/yuin/gopher-lua"
)

var (
	luaDir string
	port   int
)

// ProtectedLuaState represents a mutex-protected Lua state
type ProtectedLuaState struct {
	L    *lua.LState
	mu   sync.Mutex
	path string
}

// LuaPool manages a pool of Lua interpreters
type LuaPool struct {
	states []*ProtectedLuaState
	mu     sync.Mutex
}

func NewLuaPool(size int, luaDir string) *LuaPool {
	pool := &LuaPool{
		states: make([]*ProtectedLuaState, size),
	}

	for i := 0; i < size; i++ {
		L := lua.NewState()
		pool.states[i] = &ProtectedLuaState{L: L, path: luaDir}
		loadLuaFiles(L, luaDir)
	}

	return pool
}

func (p *LuaPool) Get() *ProtectedLuaState {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, state := range p.states {
		if state.mu.TryLock() {
			return state
		}
	}

	// If all states are busy, wait for one to become available
	state := p.states[0]
	state.mu.Lock()
	return state
}

func (p *LuaPool) Put(state *ProtectedLuaState) {
	state.mu.Unlock()
}

func (p *LuaPool) ReloadAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, state := range p.states {
		state.mu.Lock()
		loadLuaFiles(state.L, state.path)
		state.mu.Unlock()
	}
}

func (p *LuaPool) ReloadFile(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, state := range p.states {
		state.mu.Lock()
		loadLuaFile(state.L, path)
		state.mu.Unlock()
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "lua-serve",
		Short: "Serve Lua files as HTTP endpoints",
		Run:   run,
	}

	rootCmd.Flags().StringVarP(&luaDir, "dir", "d", "./lua", "Directory containing Lua files")
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to serve on")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	pool := NewLuaPool(1, luaDir)
	pool.ReloadAll()

	w := watcher.NewWatcher(
		watcher.WithPaths(luaDir),
		watcher.WithMask("**/*.lua"),
		watcher.WithWriteCallback(func(path string) error {
			fmt.Printf("Reloading Lua file: %s\n", path)
			pool.ReloadFile(path)
			return nil
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := w.Run(ctx); err != nil {
			fmt.Printf("Watcher error: %v\n", err)
		}
	}()

	e := echo.New()
	e.Any("/lua/*", handleLuaRequest(pool))
	e.GET("/debug", handleDebug(pool))

	fmt.Printf("Server running on http://localhost:%d\n", port)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

func loadLuaFiles(L *lua.LState, dir string) {
	// Load the HTTP API first
	apiPath := filepath.Join(dir, "http_api.lua")
	if err := L.DoFile(apiPath); err != nil {
		fmt.Printf("Error loading HTTP API: %v\n", err)
		return
	}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".lua") && path != apiPath {
			loadLuaFile(L, path)
		}
		return nil
	})
}

func loadLuaFile(L *lua.LState, path string) {
	fmt.Printf("Loading Lua file %s\n", path)
	if err := L.DoFile(path); err != nil {
		fmt.Printf("Error loading Lua file %s: %v\n", path, err)
	}
}

func handleLuaRequest(pool *LuaPool) echo.HandlerFunc {
	return func(c echo.Context) error {
		state := pool.Get()
		defer pool.Put(state)

		L := state.L
		method := strings.ToLower(c.Request().Method)
		path := strings.TrimPrefix(c.Request().URL.Path, "/lua")
		funcName := fmt.Sprintf("http_handler_%s_%s", method, strings.ReplaceAll(path[1:], "/", "_"))

		luaFunc := L.GetGlobal(funcName)
		if luaFunc == lua.LNil {
			return c.String(http.StatusNotFound, "Handler not found")
		}

		requestTable := L.NewTable()
		requestTable.RawSetString("method", lua.LString(method))
		requestTable.RawSetString("path", lua.LString(path))

		headers := L.NewTable()
		for k, v := range c.Request().Header {
			headers.RawSetString(k, lua.LString(v[0]))
		}
		requestTable.RawSetString("headers", headers)

		query := L.NewTable()
		for k, v := range c.QueryParams() {
			query.RawSetString(k, lua.LString(v[0]))
		}
		requestTable.RawSetString("query", query)

		body := L.NewTable()
		if err := c.Bind(&body); err == nil {
			requestTable.RawSetString("body", body)
		}

		params := L.NewTable()
		for _, name := range c.ParamNames() {
			params.RawSetString(name, lua.LString(c.Param(name)))
		}
		requestTable.RawSetString("params", params)

		L.SetField(requestTable, "get_cookie", L.NewFunction(func(L *lua.LState) int {
			name := L.ToString(1)
			cookie, err := c.Cookie(name)
			if err != nil {
				L.Push(lua.LNil)
				return 1
			}
			L.Push(lua.LString(cookie.Value))
			return 1
		}))

		err := L.CallByParam(lua.P{
			Fn:      luaFunc,
			NRet:    1,
			Protect: true,
		}, requestTable)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		ret := L.Get(-1)
		L.Pop(1)

		result := lua2.LuaValueToInterface(L, ret)
		response, ok := result.(map[string]interface{})
		if !ok {
			return c.String(http.StatusInternalServerError, "Invalid response format")
		}

		status := http.StatusOK
		if s, ok := response["status"].(float64); ok {
			status = int(s)
		}

		if headers, ok := response["headers"].(map[string]interface{}); ok {
			for k, v := range headers {
				c.Response().Header().Set(k, fmt.Sprint(v))
			}
		}

		if cookies, ok := response["cookies"].(map[string]interface{}); ok {
			for name, cookieData := range cookies {
				if cd, ok := cookieData.(map[string]interface{}); ok {
					cookie := &http.Cookie{Name: name}
					if value, ok := cd["value"].(string); ok {
						cookie.Value = value
					}
					if maxAge, ok := cd["max_age"].(float64); ok {
						cookie.MaxAge = int(maxAge)
					}
					if path, ok := cd["path"].(string); ok {
						cookie.Path = path
					}
					if domain, ok := cd["domain"].(string); ok {
						cookie.Domain = domain
					}
					if secure, ok := cd["secure"].(bool); ok {
						cookie.Secure = secure
					}
					if httpOnly, ok := cd["http_only"].(bool); ok {
						cookie.HttpOnly = httpOnly
					}
					c.SetCookie(cookie)
				}
			}
		}

		body_ := response["body"]
		switch v := body_.(type) {
		case string:
			return c.String(status, v)
		case map[string]interface{}:
			return c.JSON(status, v)
		default:
			return c.String(status, fmt.Sprintf("%v", v))
		}
	}
}

func handleDebug(pool *LuaPool) echo.HandlerFunc {
	return func(c echo.Context) error {
		state := pool.Get()
		defer pool.Put(state)

		L := state.L
		globals, ok := L.GetGlobal("_G").(*lua.LTable)
		if !ok {
			return c.String(http.StatusInternalServerError, "Failed to get _G")
		}

		functions := []string{}

		L.ForEach(globals, func(key, value lua.LValue) {
			if _, ok := value.(*lua.LFunction); ok {
				functions = append(functions, key.String())
			}
		})

		return c.JSON(http.StatusOK, map[string]interface{}{
			"registered_functions": functions,
		})
	}
}
