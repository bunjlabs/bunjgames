package common

import (
	"encoding/xml"
	"os"
	"strings"
)

// XMLElement provides a generic tree representation for navigating XML documents,
// similar to Python's ElementTree, matching elements by local name regardless of namespace.
type XMLElement struct {
	XMLName  xml.Name
	Attrs    []xml.Attr   `xml:",any,attr"`
	Content  string       `xml:",chardata"`
	Children []XMLElement `xml:",any"`
}

func ParseXMLFile(filename string) (*XMLElement, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var root XMLElement
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	return &root, nil
}

func (e *XMLElement) Find(name string) *XMLElement {
	if e == nil {
		return nil
	}
	for i := range e.Children {
		if e.Children[i].XMLName.Local == name {
			return &e.Children[i]
		}
	}
	return nil
}

func (e *XMLElement) FindAll(name string) []*XMLElement {
	if e == nil {
		return nil
	}
	var result []*XMLElement
	for i := range e.Children {
		if e.Children[i].XMLName.Local == name {
			result = append(result, &e.Children[i])
		}
	}
	return result
}

func (e *XMLElement) Text() string {
	if e == nil {
		return ""
	}
	return strings.TrimSpace(e.Content)
}

func (e *XMLElement) FindText(name string) string {
	child := e.Find(name)
	if child == nil {
		return ""
	}
	return child.Text()
}

func (e *XMLElement) Attr(name string) string {
	if e == nil {
		return ""
	}
	for _, a := range e.Attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}
