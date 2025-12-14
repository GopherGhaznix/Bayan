package hugo

import (
	"bytes"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

type MDFile struct {
	MetaData map[string]interface{}
	Body     string
	Format   string // "yaml" or "toml"
}

// ParseMD parses the file content into MDFile struct
func ParseMD(content string) (*MDFile, error) {
	// Check for TOML
	if strings.HasPrefix(content, "+++") {
		parts := strings.SplitN(content, "+++", 3)
		if len(parts) < 3 {
			// Empty front matter?? or broken
			// Only header, no body?
			return &MDFile{Body: content, Format: "toml", MetaData: make(map[string]interface{})}, nil
		}

		fm := make(map[string]interface{})
		if _, err := toml.Decode(parts[1], &fm); err != nil {
			return nil, err
		}
		return &MDFile{
			MetaData: fm,
			Body:     strings.TrimPrefix(parts[2], "\n"),
			Format:   "toml",
		}, nil
	}

	// Check for YAML
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) < 3 {
			return &MDFile{Body: content, Format: "yaml", MetaData: make(map[string]interface{})}, nil
		}

		fm := make(map[string]interface{})
		if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
			return nil, err
		}
		return &MDFile{
			MetaData: fm,
			Body:     strings.TrimPrefix(parts[2], "\n"),
			Format:   "yaml",
		}, nil
	}

	// Default/No Front Matter
	return &MDFile{Body: content, Format: "yaml", MetaData: make(map[string]interface{})}, nil
}

// ToString reconstructs the file content
func (m *MDFile) ToString() (string, error) {
	if m.Format == "toml" {
		buf := new(bytes.Buffer)
		if err := toml.NewEncoder(buf).Encode(m.MetaData); err != nil {
			return "", err
		}
		return "+++\n" + buf.String() + "+++\n" + m.Body, nil
	}

	// Default YAML
	yamlBytes, err := yaml.Marshal(m.MetaData)
	if err != nil {
		return "", err
	}
	return "---\n" + string(yamlBytes) + "---\n" + m.Body, nil
}
