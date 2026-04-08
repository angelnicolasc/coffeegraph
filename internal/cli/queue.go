package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/coffeegraph/coffeegraph/internal/config"
	"github.com/coffeegraph/coffeegraph/internal/project"
	"github.com/coffeegraph/coffeegraph/internal/queue"
)

// RunQueueAdd adds a task to the queue (via flags or interactive prompts).
func RunQueueAdd(skillFlag, taskFlag string, priorityFlag int) error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}

	skill := strings.TrimSpace(skillFlag)
	task := strings.TrimSpace(taskFlag)
	priority := priorityFlag

	// Interactive mode if flags are missing.
	if skill == "" || task == "" {
		r := bufio.NewReader(os.Stdin)
		if skill == "" {
			fmt.Printf("Which skill? (%s): ", strings.Join(config.KnownSkills, " / "))
			line, _ := r.ReadString('\n')
			skill = strings.TrimSpace(line)
		}
		if task == "" {
			fmt.Print("Task: ")
			line, _ := r.ReadString('\n')
			task = strings.TrimSpace(line)
		}
		if priority == 0 {
			fmt.Print("Priority (1-5, default 3): ")
			line, _ := r.ReadString('\n')
			line = strings.TrimSpace(line)
			if line != "" {
				if p, perr := strconv.Atoi(line); perr == nil {
					priority = p
				}
			}
		}
	}
	if priority == 0 {
		priority = 3
	}
	if skill == "" || task == "" {
		return fmt.Errorf("skill and task are required")
	}

	// Validate skill exists in project.
	skillDir := filepath.Join(root, "skills", skill)
	if _, serr := os.Stat(filepath.Join(skillDir, "SKILL.md")); serr != nil {
		return fmt.Errorf("skill %q not installed. Run: coffeegraph add %s", skill, skill)
	}

	it := queue.Item{Skill: skill, Task: task, Priority: priority}
	pos, total, err := queue.Add(root, it)
	if err != nil {
		return err
	}
	fmt.Printf("✓ Task added to %s queue (position %d of %d)\n", skill, pos, total)
	return nil
}

// RunQueueList displays all tasks in the queue.
func RunQueueList() error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	items, err := queue.Read(root)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Println("(empty queue)")
		return nil
	}
	fmt.Printf("%-4s %-18s %-4s %s\n", "#", "SKILL", "PRI", "TASK")
	fmt.Println(strings.Repeat("─", 70))
	for i, it := range items {
		task := it.Task
		if len(task) > 45 {
			task = task[:42] + "..."
		}
		fmt.Printf("%-4d %-18s P%-3d %s\n", i+1, it.Skill, it.Priority, task)
	}
	return nil
}

// RunQueueClear empties the queue.
func RunQueueClear() error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}
	if err := queue.Set(root, nil); err != nil {
		return err
	}
	fmt.Println("✓ Queue cleared.")
	return nil
}
