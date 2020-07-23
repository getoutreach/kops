/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubemanifest

import (
	"bytes"
	"fmt"

	"github.com/ghodss/yaml"
	"k8s.io/klog"
	"k8s.io/kops/util/pkg/text"
)

// Object holds arbitrary untyped kubernetes objects; it is used when we don't have the type definitions for them
type Object struct {
	data map[string]interface{}
}

// LoadObjectsFrom parses multiple objects from a yaml file
func LoadObjectsFrom(contents []byte) ([]*Object, error) {
	var objects []*Object

	// TODO: Support more separators?
	sections := text.SplitContentToSections(contents)

	for _, section := range sections {
		data := make(map[string]interface{})
		err := yaml.Unmarshal(section, &data)
		if err != nil {
			return nil, fmt.Errorf("error parsing yaml: %v", err)
		}

		obj := &Object{
			//bytes: section,
			data: data,
		}
		objects = append(objects, obj)
	}

	return objects, nil
}

// ToYAML serializes a list of objects back to bytes; it is the opposite of LoadObjectsFrom
func ToYAML(objects []*Object) ([]byte, error) {
	var yamlSeparator = []byte("\n---\n\n")
	var yamls [][]byte
	for _, object := range objects {
		// Don't serialize empty objects - they confuse yaml parsers
		if object.IsEmptyObject() {
			continue
		}

		y, err := object.ToYAML()
		if err != nil {
			return nil, fmt.Errorf("error re-marshaling manifest: %v", err)
		}

		yamls = append(yamls, y)
	}

	return bytes.Join(yamls, yamlSeparator), nil
}

func (m *Object) ToYAML() ([]byte, error) {
	b, err := yaml.Marshal(m.data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling manifest to yaml: %v", err)
	}
	return b, nil
}

func (m *Object) accept(visitor Visitor) error {
	err := visit(visitor, m.data, []string{}, func(v interface{}) {
		klog.Fatal("cannot mutate top-level data")
	})
	return err
}

// IsEmptyObject checks if the object has no keys set (i.e. `== {}`)
func (m *Object) IsEmptyObject() bool {
	return len(m.data) == 0
}

// Kind returns the kind field of the object, or "" if it cannot be found or is invalid
func (m *Object) Kind() string {
	v, found := m.data["kind"]
	if !found {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// APIVersion returns the apiVersion field of the object, or "" if it cannot be found or is invalid
func (m *Object) APIVersion() string {
	v, found := m.data["apiVersion"]
	if !found {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
