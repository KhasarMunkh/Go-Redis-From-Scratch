package command

import (
	"strings"
	"github.com/KhasarMunkh/Go-Redis-From-Scratch/resp"
	"github.com/KhasarMunkh/Go-Redis-From-Scratch/storage"
)

// Handlers depends on the storage interface
type Ctx struct {
	Storage storage.Storage
}

// Handler is a function that takes a context and a command and returns a response
type Handler func(ctx *Ctx, args []resp.Value) resp.Value 

type Registry struct {
	commands map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Handler),
	}
}

func (r *Registry) Register(key string, h Handler) {
	r.commands[strings.ToUpper(key)] = h
}

func (r *Registry) Exec(ctx *Ctx, cmd string, args []resp.Value) resp.Value {
	if h, ok := r.commands[cmd]; ok {
		return h(ctx, args)
	}
	return resp.NewError("Error: Unknown command " + cmd + ".")
}

// RegisterBasic registers the basic commands like GET, SET, DEL, PING
func RegisterBasic(r *Registry) {
	r.Register("GET", get)
	r.Register("SET", set)
	r.Register("DEL", del)
	r.Register("PING", ping)
}

func argToString(v resp.Value) string {
	if v.Typ == "bulk" {
		return v.Bulk
	}
	if v.Typ == "string" {
		return v.Str
	}
	return ""
}

func get(ctx *Ctx, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.NewError("ERR: Wrong number of args for GET")
	}
	key := argToString(args[0])
	if v, ok := ctx.Storage.Get(key); ok {
		return resp.NewBulkString(v)
	} else {
		return resp.NewNull()
	}
}

func set(ctx *Ctx, args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.NewError("ERR: Wrong number of args for SET")
	}
	key := argToString(args[0])
	val := argToString(args[1])
	ctx.Storage.Set(key, val)
	return resp.NewSimpleString("OK")
}

func del(ctx *Ctx, args []resp.Value) resp.Value {
	if len(args) < 1 {
		return resp.NewError("ERR: Wrong number of args for DEL")
	}
	keys := make([]string, len(args))
	for i, a := range args {
		switch a.Typ {
		case "bulk":
			keys[i] = a.Bulk
		case "string":
			keys[i] = a.Bulk
		default: 
			return resp.NewError("key string error")
		}
	}
	n := ctx.Storage.Delete(keys...)
	return resp.NewInteger(n)
}

func ping(_ *Ctx, args []resp.Value) resp.Value {
	if len(args) > 1 {
		return resp.NewError("ERR: Wrong number of args for PING")
	}
	if len(args) == 0 {
		return resp.NewSimpleString("PONG")
	}
	return resp.NewSimpleString(argToString(args[0]))
}
