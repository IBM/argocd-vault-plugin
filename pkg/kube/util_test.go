package kube

import (
	"errors"
	"reflect"
	"testing"
)

func assertSuccessfulReplacement(actual, expected *Resource, t *testing.T) {
	if len(actual.replacementErrors) != 0 {
		t.Fatalf("expected 0 errors but got: %s", actual.replacementErrors)
	}

	if !reflect.DeepEqual(actual.TemplateData, expected.TemplateData) {
		t.Fatalf("expected replaced template to look like %s\n but got: %s", expected.TemplateData, actual.TemplateData)
	}

	if !reflect.DeepEqual(actual.VaultData, expected.VaultData) {
		t.Fatalf("expected Vault map to look like %s\n but got: %s", expected.VaultData, actual.VaultData)
	}
}

func assertFailedReplacement(actual, expected *Resource, t *testing.T) {
	if !reflect.DeepEqual(actual.replacementErrors, expected.replacementErrors) {
		t.Fatalf("expected replacementErrors: %s but got %s", expected.replacementErrors, actual.replacementErrors)
	}

	if !reflect.DeepEqual(actual.TemplateData, expected.TemplateData) {
		t.Fatalf("expected replaced template to look like %s\n but got: %s", expected.TemplateData, actual.TemplateData)
	}

	if !reflect.DeepEqual(actual.VaultData, expected.VaultData) {
		t.Fatalf("expected Vault map to look like %s\n but got: %s", expected.VaultData, actual.VaultData)
	}
}

func TestGenericReplacement_simpleString(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace>",
		},
		VaultData: map[string]interface{}{
			"namespace": "default",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "default",
		},
		VaultData: map[string]interface{}{
			"namespace": "default",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_multiString(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace>",
			"image":     "foo.io/<name>:<tag>",
		},
		VaultData: map[string]interface{}{
			"namespace": "default",
			"name":      "app",
			"tag":       "latest",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "default",
			"image":     "foo.io/app:latest",
		},
		VaultData: map[string]interface{}{
			"namespace": "default",
			"name":      "app",
			"tag":       "latest",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_nestedString(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace>",
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app": "<name>",
				},
			},
		},
		VaultData: map[string]interface{}{
			"namespace": "default",
			"name":      "foo",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "default",
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"app": "foo",
				},
			},
		},
		VaultData: map[string]interface{}{
			"namespace": "default",
			"name":      "foo",
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_int(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace>",
			"spec": map[string]interface{}{
				"replicas": "<replicas>",
			},
		},
		VaultData: map[string]interface{}{
			"namespace": "default",
			"replicas":  1,
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "default",
			"spec": map[string]interface{}{
				"replicas": 1,
			},
		},
		VaultData: map[string]interface{}{
			"namespace": "default",
			"replicas":  1,
		},
		replacementErrors: []error{},
	}

	assertSuccessfulReplacement(&dummyResource, &expected, t)
}

func TestGenericReplacement_missingValue(t *testing.T) {
	dummyResource := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "<namespace>",
			"spec": map[string]interface{}{
				"replicas": "<replicas>",
			},
		},
		VaultData: map[string]interface{}{
			"namespace": "default",
		},
	}

	replaceInner(&dummyResource, &dummyResource.TemplateData, genericReplacement)

	expected := Resource{
		TemplateData: map[string]interface{}{
			"namespace": "default",
			"spec": map[string]interface{}{
				"replicas": "<replicas>",
			},
		},
		VaultData: map[string]interface{}{
			"namespace": "default",
		},
		replacementErrors: []error{
			errors.New("replaceString: missing Vault value for placeholder replicas in string replicas: <replicas>"),
		},
	}

	assertFailedReplacement(&dummyResource, &expected, t)
}
