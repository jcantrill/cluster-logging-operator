package parser

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"regexp"
	"strings"
)

type ExperimentalCLFTuning struct {
	Inputs  []ExperimentalSectionParameters `json:"inputs,omitempty"`
	Outputs []ExperimentalSectionParameters `json:"outputs,omitempty"`
}

type ExperimentalSectionParameters struct {
	Name             string         `json:"name,omitempty"`
	Params           map[string]any `json:"params,omitempty"`
	TransformsParams map[string]any `json:"transformsParams,omitempty"`
}

type TomlConfig []*TomlSection

func (c TomlConfig) Modify(tuning ExperimentalCLFTuning) TomlConfig {
	return c.ModifyInputs(tuning.Inputs).ModifyOutputs(tuning.Outputs)
}

func (c TomlConfig) ModifyInputs(mods []ExperimentalSectionParameters) TomlConfig {

	for _, mod := range mods {
		inputID := "sources." + helpers.MakeInputID(mod.Name)
		log.V(4).Info("trying to find section", "id", inputID)
		inputRE := regexp.MustCompile(`^sources.input_[a-z]*_[a-z]*$`)
		if section := c.FindSection(func(sectionID string) bool {
			return inputRE.MatchString(sectionID) && strings.HasPrefix(sectionID, inputID)
		}); section != nil {
			for k, v := range mod.Params {
				log.V(4).Info("modifying", "section", section.ID, "key", k, "value", v)
				section.Config[k] = v
			}
		} else {
			log.V(4).Info("section not found", "inputID", inputID)
		}

	}

	return c
}
func (c TomlConfig) ModifyOutputs(mods []ExperimentalSectionParameters) TomlConfig {

	for _, mod := range mods {
		id := helpers.MakeOutputID(mod.Name)
		sectionID := "sinks." + id
		log.V(4).Info("trying to find section", "id", sectionID)
		inputRE := regexp.MustCompile(fmt.Sprintf(`^%s$`, sectionID))
		if section := c.FindSection(func(sectionID string) bool {
			return inputRE.MatchString(sectionID) && strings.HasPrefix(sectionID, sectionID)
		}); section != nil {
			for k, v := range mod.Params {
				log.V(4).Info("modifying", "section", section.ID, "key", k, "value", v)
				section.Config[k] = v
			}
		} else {
			log.V(0).Info("base section not found", "sectionID", sectionID)
		}

		log.V(0).Info("checking output transforms", "fields", mod.TransformsParams)
		if len(mod.TransformsParams) > 0 {

			for k, v := range mod.TransformsParams {
				var post string
				if bits := strings.Split(k, "/"); len(bits) > 1 {
					k = bits[0]
					post = bits[1]
				}

				sectionRE := regexp.MustCompile(fmt.Sprintf(`^transforms.%s.*%s.*`, id, post))
				log.V(0).Info("looking for section", "regexp", sectionRE)
				if section := c.FindSection(func(sectionID string) bool {
					log.V(0).Info("checking", "id", sectionID)
					return sectionRE.MatchString(sectionID)
				}); section != nil {
					section.Config[k] = v
				}

			}

			if section := c.FindSection(func(sectionID string) bool {
				return inputRE.MatchString(sectionID) && strings.HasPrefix(sectionID, sectionID)
			}); section != nil {
				for k, v := range mod.Params {
					log.V(4).Info("modifying", "section", section.ID, "key", k, "value", v)
					section.Config[k] = v
				}
			} else {
				log.V(0).Info("base section not found", "sectionID", sectionID)
			}
		}
	}

	return c
}

func (c TomlConfig) FindSection(match func(string) bool) *TomlSection {
	for _, section := range []*TomlSection(c) {
		log.V(4).Info("Checking section", "id", section.ID)
		if match(section.ID) {
			return section
		}
	}
	return nil
}

func (c TomlConfig) String() string {
	var toml []string
	for _, section := range c {
		toml = append(toml, section.String())
	}
	return strings.Join(toml, "\n\n")
}

type TomlSection struct {
	ID     string
	Config map[string]any
}

func NewTomlSection(id string) *TomlSection {
	return &TomlSection{
		ID:     id,
		Config: map[string]any{},
	}
}

func (t TomlSection) String() string {
	var out []string
	if t.ID != "" {
		out = append(out, fmt.Sprintf("[%s]", t.ID))
	}
	for k, v := range t.Config {
		out = append(out, fmt.Sprintf("%s = %v", k, v))
	}
	return strings.Join(out, "\n")
}

func ParseToml(s string) (results TomlConfig) {

	keyValueRE := regexp.MustCompile(`^(?P<k>[a-zA-z_]*)\s*=\s*(?P<v>.*)$`)
	keyId := keyValueRE.SubexpIndex("k")
	valueId := keyValueRE.SubexpIndex("v")

	headerRE := regexp.MustCompile(`^\[(?P<id>.*)\]$`)
	idIndex := headerRE.SubexpIndex("id")

	sourceStartRE := regexp.MustCompile(`^source\s*=\s*'''`)
	sourceEndRE := regexp.MustCompile(`.*'''$`)

	current := NewTomlSection("")
	results = append(results, current)
	var sourceLines []string
	for _, line := range strings.Split(s, "\n") {
		if sourceStartRE.MatchString(line) {
			sourceLines = []string{`'''`}
		} else if sourceEndRE.MatchString(line) {
			sourceLines = append(sourceLines, line)
			current.Config["source"] = strings.Join(sourceLines, "\n")
			sourceLines = []string{}
		} else if len(sourceLines) > 0 {
			sourceLines = append(sourceLines, line)
		} else if matches := headerRE.FindStringSubmatch(line); len(matches) > 0 {
			current = NewTomlSection(matches[idIndex])
			results = append(results, current)
		} else if matches = keyValueRE.FindStringSubmatch(line); len(matches) > 0 {
			current.Config[matches[keyId]] = matches[valueId]
		}
	}
	return results
}
