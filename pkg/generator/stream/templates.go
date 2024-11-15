package stream

import (
	_ "embed"
	"log"
	"text/template"
)

//go:embed _serializer.c.tpl
var tpl string

var serializerTemplate *template.Template

// SerializerTemplateData is the data passed to the serializer template
type SerializerTemplateData struct {
	ProtoMessageName string
	ProtoPackageName string
}

func init() {
	var err error

	serializerTemplate, err = template.New("serializerTemplate").Funcs(nil).Parse(tpl)

	if err != nil {
		log.Fatal(err)
	}
}
