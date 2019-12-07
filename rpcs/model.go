//author: richard
package rpcs

import "github.com/advancevillage/3rd/logs"

//@link: https://golang.org/pkg/net/rpc/
//- the method's type is exported.
//- the method is exported.
//- the method has two arguments, both exported (or builtin) types.
//- the method's second argument is a pointer.
//- the method has return type error.
type Server struct {
	host   string
	port   int
	logger logs.Logs
	rcvr   []interface{}
}
