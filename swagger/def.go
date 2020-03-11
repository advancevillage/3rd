//author: richard
package swagger


type ISwagger struct {
	Swagger  string `json:"swagger,omitempty"`
	Info     *IInfo  `json:"info,omitempty"`
	Host     string  `json:"host,omitempty"`
	Paths    map[string]map[string]IMethod `json:"paths,omitempty"`
	Definitions map[string]IDefinitions `json:"definitions,omitempty"`
}


type IInfo struct {
	Description  string `json:"description,omitempty"`
	Title        string `json:"title,omitempty"`
	Contact      struct{
		Name  string `json:",omitempty"`
		Email string `json:"email,omitempty"`
	} `json:"contact,omitempty"`
	License      struct{
		Name  string `json:"name,omitempty"`
		URL   string `json:"url,omitempty"`
	} `json:"license,omitempty"`
	Version  string `json:"version,omitempty"`
}

type IDefinitions struct {
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type IMethod struct {
	Produces []string  `json:"produces,omitempty"`
	Summary  string    `json:"summary,omitempty"`
	Parameters []struct{
		Type string `json:"type,omitempty"`
		Description string `json:"description,omitempty"`
		Name string `json:"name,omitempty"`
		In   string `json:"in,omitempty"`
		Required bool `json:"required,omitempty"`
	} `json:"parameters,omitempty"`
	Responses map[string]struct{
		Schema map[string]interface{}
	} `json:"responses,omitempty"`
}