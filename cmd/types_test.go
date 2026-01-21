package cmd

import (
	"testing"
)

func TestModuleDirs(t *testing.T) {
	expected := []string{"components", "bases", "projects"}

	if len(ModuleDirs) != len(expected) {
		t.Fatalf("expected %d module dirs, got %d", len(expected), len(ModuleDirs))
	}

	for i, dir := range expected {
		if ModuleDirs[i] != dir {
			t.Errorf("ModuleDirs[%d] = '%s', expected '%s'", i, ModuleDirs[i], dir)
		}
	}
}

func TestModuleTypeOrder(t *testing.T) {
	// Components should sort first, then bases, then projects
	if ModuleTypeOrder[TypeComponent] >= ModuleTypeOrder[TypeBase] {
		t.Error("components should sort before bases")
	}
	if ModuleTypeOrder[TypeBase] >= ModuleTypeOrder[TypeProject] {
		t.Error("bases should sort before projects")
	}
}

func TestDirConstants(t *testing.T) {
	if DirComponents != "components" {
		t.Errorf("DirComponents = '%s', expected 'components'", DirComponents)
	}
	if DirBases != "bases" {
		t.Errorf("DirBases = '%s', expected 'bases'", DirBases)
	}
	if DirProjects != "projects" {
		t.Errorf("DirProjects = '%s', expected 'projects'", DirProjects)
	}
}

func TestTypeConstants(t *testing.T) {
	if TypeComponent != "component" {
		t.Errorf("TypeComponent = '%s', expected 'component'", TypeComponent)
	}
	if TypeBase != "base" {
		t.Errorf("TypeBase = '%s', expected 'base'", TypeBase)
	}
	if TypeProject != "project" {
		t.Errorf("TypeProject = '%s', expected 'project'", TypeProject)
	}
}
