package httpie

import "strings"

type CmdLine struct {
	Flags   []*Flag
	Method  *Method
	URL     string
	Items   []*Item
	HasBody bool // has body
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

	s = append(s, cl.Method.String())
	s = append(s, string(cl.URL))

	// items
	for _, v := range cl.Items {
		s = append(s, v.String())
	}

	return strings.Join(s, " ")
}

func NewCmdLine() *CmdLine {
	return &CmdLine{
		Flags:  make([]*Flag, 0),
		Items:  make([]*Item, 0),
		Method: NewMethod(""),
	}
}