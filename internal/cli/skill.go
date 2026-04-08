package cli

import (
	"fmt"
	"os/exec"

	"github.com/coffeegraph/coffeegraph/internal/project"
	"github.com/coffeegraph/coffeegraph/internal/registry"
)

// RunSkillInstall installs one community skill into current project.
func RunSkillInstall(name string) error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	dst, err := registry.Install(root, name)
	if err != nil {
		return err
	}
	fmt.Println("Installed skill at", dst)
	return nil
}

// RunSkillList lists available community skills.
func RunSkillList() error {
	items, err := registry.List()
	if err != nil {
		return err
	}
	for _, it := range items {
		fmt.Println(it)
	}
	return nil
}

// RunSkillPublish opens the community registry PR template page.
func RunSkillPublish() error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	skills, err := registry.LocalCustomSkills(root)
	if err != nil {
		return err
	}
	url := registry.PublishTemplateURL(skills)
	fmt.Println("Open this URL to publish your skills:")
	fmt.Println(url)
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	return nil
}
