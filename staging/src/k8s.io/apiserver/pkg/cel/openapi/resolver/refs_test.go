/*
Copyright 2025 The Kubernetes Authors.

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

package resolver

import (
	"testing"

	"k8s.io/kube-openapi/pkg/validation/spec"
)

func TestPopulateRefs_DoesNotMutateOriginalItems(t *testing.T) {
	// Create a schema with Items that has a Ref
	innerSchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"string"},
		},
	}

	itemsSchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Ref: spec.MustCreateRef("#/definitions/Inner"),
		},
	}

	arraySchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"array"},
			Items: &spec.SchemaOrArray{
				Schema: itemsSchema,
			},
		},
	}

	// Store original pointer to verify it wasn't mutated
	originalItemsSchema := arraySchema.Items.Schema

	schemas := map[string]*spec.Schema{
		"#/definitions/Array": arraySchema,
		"#/definitions/Inner": innerSchema,
	}

	schemaOf := func(ref string) (*spec.Schema, bool) {
		s, ok := schemas[ref]
		return s, ok
	}

	result, err := PopulateRefs(schemaOf, "#/definitions/Array")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the result Items.Schema was populated (resolved the ref)
	if result.Items == nil || result.Items.Schema == nil {
		t.Fatal("expected result to have Items.Schema")
	}
	if len(result.Items.Schema.Type) != 1 || result.Items.Schema.Type[0] != "string" {
		t.Errorf("expected Items.Schema to be resolved to string type, got %v", result.Items.Schema.Type)
	}

	// Critical: verify the original was NOT mutated
	if arraySchema.Items.Schema != originalItemsSchema {
		t.Error("original arraySchema.Items.Schema pointer was mutated")
	}
	if arraySchema.Items.Schema.Ref.String() != "#/definitions/Inner" {
		t.Errorf("original arraySchema.Items.Schema.Ref was mutated, got %v", arraySchema.Items.Schema.Ref.String())
	}

	// Verify result has a different Items pointer than original
	if result.Items == arraySchema.Items {
		t.Error("result.Items should be a different pointer than original arraySchema.Items")
	}
}

func TestPopulateRefs_DoesNotMutateOriginalAdditionalProperties(t *testing.T) {
	// Create a schema with AdditionalProperties that has a Ref
	innerSchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"integer"},
		},
	}

	additionalPropsSchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Ref: spec.MustCreateRef("#/definitions/Inner"),
		},
	}

	mapSchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			AdditionalProperties: &spec.SchemaOrBool{
				Allows: true,
				Schema: additionalPropsSchema,
			},
		},
	}

	// Store original pointer to verify it wasn't mutated
	originalAdditionalPropsSchema := mapSchema.AdditionalProperties.Schema

	schemas := map[string]*spec.Schema{
		"#/definitions/Map":   mapSchema,
		"#/definitions/Inner": innerSchema,
	}

	schemaOf := func(ref string) (*spec.Schema, bool) {
		s, ok := schemas[ref]
		return s, ok
	}

	result, err := PopulateRefs(schemaOf, "#/definitions/Map")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the result AdditionalProperties.Schema was populated (resolved the ref)
	if result.AdditionalProperties == nil || result.AdditionalProperties.Schema == nil {
		t.Fatal("expected result to have AdditionalProperties.Schema")
	}
	if len(result.AdditionalProperties.Schema.Type) != 1 || result.AdditionalProperties.Schema.Type[0] != "integer" {
		t.Errorf("expected AdditionalProperties.Schema to be resolved to integer type, got %v", result.AdditionalProperties.Schema.Type)
	}

	// Critical: verify the original was NOT mutated
	if mapSchema.AdditionalProperties.Schema != originalAdditionalPropsSchema {
		t.Error("original mapSchema.AdditionalProperties.Schema pointer was mutated")
	}
	if mapSchema.AdditionalProperties.Schema.Ref.String() != "#/definitions/Inner" {
		t.Errorf("original mapSchema.AdditionalProperties.Schema.Ref was mutated, got %v", mapSchema.AdditionalProperties.Schema.Ref.String())
	}

	// Verify result has a different AdditionalProperties pointer than original
	if result.AdditionalProperties == mapSchema.AdditionalProperties {
		t.Error("result.AdditionalProperties should be a different pointer than original mapSchema.AdditionalProperties")
	}
}

func TestPopulateRefs_DoesNotMutateWhenNoChanges(t *testing.T) {
	// Create a schema without any refs - should return the same pointer
	schema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"string"},
		},
	}

	schemas := map[string]*spec.Schema{
		"#/definitions/Simple": schema,
	}

	schemaOf := func(ref string) (*spec.Schema, bool) {
		s, ok := schemas[ref]
		return s, ok
	}

	result, err := PopulateRefs(schemaOf, "#/definitions/Simple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// When no changes are needed, the same pointer should be returned
	if result != schema {
		t.Error("expected same pointer when no changes needed")
	}
}

func TestPopulateRefs_DoesNotMutateNestedItems(t *testing.T) {
	// Test nested array (array of arrays with refs)
	innerSchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"number"},
		},
	}

	innerItemsSchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Ref: spec.MustCreateRef("#/definitions/Inner"),
		},
	}

	innerArraySchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"array"},
			Items: &spec.SchemaOrArray{
				Schema: innerItemsSchema,
			},
		},
	}

	outerArraySchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"array"},
			Items: &spec.SchemaOrArray{
				Schema: innerArraySchema,
			},
		},
	}

	originalInnerItems := innerArraySchema.Items
	originalInnerItemsSchema := innerArraySchema.Items.Schema

	schemas := map[string]*spec.Schema{
		"#/definitions/OuterArray": outerArraySchema,
		"#/definitions/Inner":      innerSchema,
	}

	schemaOf := func(ref string) (*spec.Schema, bool) {
		s, ok := schemas[ref]
		return s, ok
	}

	_, err := PopulateRefs(schemaOf, "#/definitions/OuterArray")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Critical: verify the original nested schema was NOT mutated
	if innerArraySchema.Items != originalInnerItems {
		t.Error("original innerArraySchema.Items pointer was mutated")
	}
	if innerArraySchema.Items.Schema != originalInnerItemsSchema {
		t.Error("original innerArraySchema.Items.Schema pointer was mutated")
	}
	if innerArraySchema.Items.Schema.Ref.String() != "#/definitions/Inner" {
		t.Errorf("original innerArraySchema.Items.Schema.Ref was mutated, got %v", innerArraySchema.Items.Schema.Ref.String())
	}
}
