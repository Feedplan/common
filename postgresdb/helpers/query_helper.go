package query_helpers

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

func GetQueryFromTemplate(queriesFS embed.FS, templateFileName string, data any, funcMap *template.FuncMap) (*string, error) {
	tpl, err := queriesFS.ReadFile(templateFileName)
	if err != nil {
		fmt.Printf("error reading template : %v\n", err)
		return nil, err
	}

	if err != nil {
		fmt.Printf("error parsing FS: %v", err)
		return nil, err
	}

	tmpl := template.New(templateFileName)
	if funcMap != nil {
		tmpl = tmpl.Funcs(*funcMap)
	}

	tmpl, err = tmpl.Parse(string(tpl))
	if err != nil {
		fmt.Printf("error parsing template: %v\n", err)
		return nil, err
	}

	var sql bytes.Buffer
	err = tmpl.Execute(&sql, data)
	if err != nil {
		fmt.Printf("error executing template: %v\n", err)
		return nil, err
	}

	sqlStr := sql.String()
	return &sqlStr, nil
}
