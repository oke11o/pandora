package scenario

import (
	"fmt"
	"io"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/spf13/afero"

	"github.com/yandex/pandora/components/providers/scenario/postprocessor"
)

type AmmoHCL struct {
	Variables       map[string]string `hcl:"variables"`
	VariableSources []SourceHCL       `hcl:"variable_source,block"`
	Requests        []RequestHCL      `hcl:"request,block"`
	Scenarios       []ScenarioHCL     `hcl:"scenario,block"`
}

type SourceHCL struct {
	Name           string   `hcl:"name,label"`
	Type           string   `hcl:"type,label"`
	File           string   `hcl:"file"`
	Fields         []string `hcl:"fields"`
	SkipHeader     bool     `hcl:"skip_header"`
	HeaderAsFields bool     `hcl:"header_as_fields"`
}

type RequestHCL struct {
	Name           string             `hcl:"name,label"`
	Method         string             `hcl:"method"`
	Headers        map[string]string  `hcl:"headers"`
	Tag            string             `hcl:"tag"`
	Body           *string            `hcl:"body"`
	Uri            string             `hcl:"uri"`
	Preprocessor   PreprocessorHCL    `hcl:"preprocessor,block"`
	Postprocessors []PostprocessorHCL `hcl:"postprocessor,block"`
	Templater      string             `hcl:"templater"`
}

type ScenarioHCL struct {
	Name           string   `hcl:"name,label"`
	Weight         int64    `hcl:"weight"`
	MinWaitingTime int64    `hcl:"min_waiting_time"`
	Shoots         []string `hcl:"shoot"`
}

type PostprocessorHCL struct {
	Type    string            `hcl:"type,label"`
	Mapping map[string]string `hcl:"mapping"`
}

type PreprocessorHCL struct {
	Name      string            `hcl:"name,label"`
	Variables map[string]string `hcl:"variables"`
}

func ParseHCLFile(file afero.File) (AmmoHCL, error) {
	const op = "hcl.ParseHCLFile"

	var config AmmoHCL
	bytes, err := io.ReadAll(file)
	if err != nil {
		return AmmoHCL{}, fmt.Errorf("%s, io.ReadAll, %w", op, err)
	}
	err = hclsimple.Decode(file.Name(), bytes, nil, &config)
	if err != nil {
		return AmmoHCL{}, fmt.Errorf("%s, hclsimple.Decode, %w", op, err)
	}
	return config, nil
}

func ConvertHCLToAmmo(ammo AmmoHCL, fs afero.Fs) (AmmoConfig, error) {
	const op = "scenario.ConvertHCLToAmmo"

	sources := make([]VariableSource, len(ammo.VariableSources))
	for i, s := range ammo.VariableSources {
		switch s.Type {
		case "file/json":
			sources[i] = &VariableSourceJson{
				Name: s.Name,
				File: s.File,
				fs:   fs,
			}
		case "file/csv":
			sources[i] = &VariableSourceCsv{
				Name:           s.Name,
				File:           s.File,
				Fields:         s.Fields,
				SkipHeader:     s.SkipHeader,
				HeaderAsFields: s.HeaderAsFields,
				fs:             fs,
			}
		default:
			return AmmoConfig{}, fmt.Errorf("%s, unknown variable source type: %s", op, s.Type)
		}
	}
	requests := make([]RequestConfig, len(ammo.Requests))
	for i, r := range ammo.Requests {
		postprocessors := make([]postprocessor.Postprocessor, len(r.Postprocessors))
		for j, p := range r.Postprocessors {
			switch p.Type {
			case "var/header":
				postprocessors[j] = &postprocessor.VarHeaderPostprocessor{
					Mapping: p.Mapping,
				}
			case "var/xpath":
				postprocessors[j] = &postprocessor.VarXpathPostprocessor{
					Mapping: p.Mapping,
				}
			case "var/jsonpath":
				postprocessors[j] = &postprocessor.VarJsonpathPostprocessor{
					Mapping: p.Mapping,
				}
			default:
				return AmmoConfig{}, fmt.Errorf("%s, unknown postprocessor type: %s", op, p.Type)
			}
		}
		requests[i] = RequestConfig{
			Name:           r.Name,
			Method:         r.Method,
			Headers:        r.Headers,
			Tag:            r.Tag,
			Body:           r.Body,
			Uri:            r.Uri,
			Preprocessor:   Preprocessor{Name: r.Preprocessor.Name, Variables: r.Preprocessor.Variables},
			Postprocessors: postprocessors,
			Templater:      r.Templater,
		}
	}

	scenarios := make([]ScenarioConfig, len(ammo.Scenarios))
	for i, s := range ammo.Scenarios {
		scenarios[i] = ScenarioConfig{
			Name:           s.Name,
			Weight:         s.Weight,
			MinWaitingTime: s.MinWaitingTime,
			Shoots:         s.Shoots,
		}
	}

	result := AmmoConfig{
		Variables:       ammo.Variables,
		VariableSources: sources,
		Requests:        requests,
		Scenarios:       scenarios,
	}

	return result, nil
}

func ConvertAmmoToHCL(ammo AmmoConfig) (AmmoHCL, error) {
	const op = "scenario.ConvertHCLToAmmo"

	sources := make([]SourceHCL, len(ammo.VariableSources))
	for i, s := range ammo.VariableSources {
		switch val := s.(type) {
		case *VariableSourceJson:
			v := SourceHCL{
				Type: "file/json",
				Name: val.Name,
				File: val.File,
			}
			sources[i] = v
		case *VariableSourceCsv:
			v := SourceHCL{
				Type:           "file/csv",
				Name:           val.Name,
				File:           val.File,
				Fields:         val.Fields,
				SkipHeader:     val.SkipHeader,
				HeaderAsFields: val.HeaderAsFields,
			}
			sources[i] = v
		default:
			return AmmoHCL{}, fmt.Errorf("%s variable source type %T not supported", op, val)
		}
	}

	requests := make([]RequestHCL, len(ammo.Requests))
	for i, r := range ammo.Requests {
		postprocessors := make([]PostprocessorHCL, len(r.Postprocessors))
		for j, p := range r.Postprocessors {
			switch val := p.(type) {
			case *postprocessor.VarHeaderPostprocessor:
				postprocessors[j] = PostprocessorHCL{
					Type:    "var/header",
					Mapping: val.Mapping,
				}
			case *postprocessor.VarXpathPostprocessor:
				postprocessors[j] = PostprocessorHCL{
					Type:    "var/xpath",
					Mapping: val.Mapping,
				}
			case *postprocessor.VarJsonpathPostprocessor:
				postprocessors[j] = PostprocessorHCL{
					Type:    "var/jsonpath",
					Mapping: val.Mapping,
				}
			default:
				return AmmoHCL{}, fmt.Errorf("%s postprocessor type %T not supported", op, val)
			}
		}

		req := RequestHCL{
			Name:           r.Name,
			Uri:            r.Uri,
			Method:         r.Method,
			Headers:        r.Headers,
			Tag:            r.Tag,
			Body:           r.Body,
			Templater:      r.Templater,
			Postprocessors: postprocessors,
			Preprocessor: PreprocessorHCL{
				Name:      r.Preprocessor.Name,
				Variables: r.Preprocessor.Variables,
			},
		}
		requests[i] = req
	}
	scenarios := make([]ScenarioHCL, len(ammo.Scenarios))
	for i, s := range ammo.Scenarios {
		scenarios[i] = ScenarioHCL{
			Name:           s.Name,
			Weight:         s.Weight,
			MinWaitingTime: s.MinWaitingTime,
			Shoots:         s.Shoots,
		}
	}

	result := AmmoHCL{
		Variables:       ammo.Variables,
		VariableSources: sources,
		Requests:        requests,
		Scenarios:       scenarios,
	}

	return result, nil
}
