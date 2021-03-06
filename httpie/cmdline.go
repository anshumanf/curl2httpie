package httpie

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type CmdLine struct {
	Flags         []*Flag
	Method        *Method
	URL           string
	Items         []*Item
	HasBody       bool
	DirectedInput io.ReadCloser
}

func (cl *CmdLine) AddFlag(f *Flag) {
	cl.Flags = append(cl.Flags, f)
}

func (cl *CmdLine) SetMethod(m *Method) {
	cl.Method = m
}

func (cl *CmdLine) SetURL(url string) {
	cl.URL = url
}

func (cl *CmdLine) AddItem(i *Item) {
	cl.Items = append(cl.Items, i)
}

func (cl *CmdLine) String() string {
	if cl.Method == nil {
		cl.Method = NewMethod("")
	}

	// slice
	s := make([]string, 0, len(cl.Flags)+len(cl.Items)+3) // http method url
	s = append(s, "http")
	// flags

	// default flag
	foundContentType := false
	for _, v := range cl.Flags {
		if v.Long == "json" || v.Long == "form" {
			foundContentType = true
		}
		s = append(s, v.String())
	}

	if !foundContentType && cl.HasBody && cl.Method.String() == "GET" {
		cl.SetMethod(NewMethod("POST")) // default post if has body
	}

	if !foundContentType && cl.Method.String() != "GET" {
		s = append(s, "--form")
	}

	s = append(s, cl.Method.String(), fmt.Sprintf("'%s'", cl.URL))

	for _, v := range cl.Items {
		s = append(s, v.String())
	}

	if cl.DirectedInput != nil && cl.HasBody {
		bytes, err := ioutil.ReadAll(cl.DirectedInput)
		if err != nil {
			fmt.Println("Skipped: Read DirectedInput error", err.Error())
		} else {
			s = append([]string{"echo", string(bytes), "|"}, s...)
			cl.DirectedInput.Close()
		}
	}

	return strings.Join(s, " ")
}

func NewCmdLine() *CmdLine {
	return &CmdLine{
		Flags: make([]*Flag, 0),
		Items: make([]*Item, 0),
	}
}

func NewCmdLineByArgs(args []string) (*CmdLine, error) {
	cmdLine := NewCmdLine()
	if len(args) == 1 {
		cmdLine.URL = args[0]
		return cmdLine, nil
	}

	var err error
	cmdLine.Flags, err = getFlagsByArgs(args)
	if err != nil {
		return nil, errors.Wrap(err, "NewCmdLineByArgs")
	}

	cmdLine.Method, cmdLine.URL, cmdLine.Items, err = getMethodURLAndItems(args)
	return cmdLine, nil
}

func getMethodURLAndItems(args []string) (method *Method, url string, items []*Item, err error) {
	method = NewMethod("")

	var lastFlagIndex int
	foundFlag := false
	for i := len(args) - 1; i >= 0; i-- {
		if strings.HasPrefix(args[i], "-") {
			lastFlagIndex = i
			foundFlag = true
			break
		}
	}

	possibleMethodIndex := 0
	if foundFlag {
		var flags []*Flag
		flags, err = getFlagsByArgs(args[lastFlagIndex:])
		if err != nil {
			return
		}
		if len(flags) < 1 {
			err = errors.New("invalid flags")
			return
		}
		if flags[0].HasArg {
			possibleMethodIndex = lastFlagIndex + 2
		} else {
			possibleMethodIndex = lastFlagIndex + 1
		}
	}

	urlIndex := possibleMethodIndex
	possibleMethod := strings.ToUpper(args[possibleMethodIndex])
	if inStringSlice(httpMethods, possibleMethod) {
		method = NewMethod(possibleMethod)
		urlIndex = possibleMethodIndex + 1
	}
	url = args[urlIndex]

	if len(args) > urlIndex+1 {
		items, err = parseItems(args[urlIndex+1:])
	}
	return
}

func parseItems(args []string) ([]*Item, error) {
	items := make([]*Item, 0, len(args))
	for _, arg := range args {
		item, err := getItemByArg(arg)
		if err != nil {
			return nil, err
		}

		if item != nil {
			items = append(items, item)
		}
	}
	return items, nil
}

func getItemByArg(arg string) (*Item, error) {
	for i, r := range arg {
		if i > 0 && arg[i-1] == '\\' {
			continue
		}

		switch r {
		case '@':
			return NewFileField(arg[:i], arg[i+1:]), nil
		case ':':
			if arg[i+1] == '=' {
				return NewJSONField(arg[:i], arg[i+2:]), nil
			}
			return NewHeader(arg[:i], arg[i+1:]), nil
		case '=':
			if arg[i+1] == '=' {
				return NewURLParam(arg[:i], arg[i+2:]), nil
			}
			return NewDataField(arg[:i], arg[i+1:]), nil
		}
	}

	return nil, errors.New("unknown item")
}

var httpMethods = []string{
	http.MethodDelete,
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

func inStringSlice(slice []string, target string) bool {
	for _, v := range slice {
		if v == target {
			return true
		}
	}

	return false
}
