package main

import (
	"fmt"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

// SkillTreeView displays combat and trade upgrades as trees.
type SkillTreeView struct {
	widgets.Component
	game       *Game
	combatTree *widgets.Tree
	tradeTree  *widgets.Tree
	focusLeft  bool

	style          backend.Style
	unlockedStyle  backend.Style
	availableStyle backend.Style
	lockedStyle    backend.Style
}

func NewSkillTreeView(game *Game) *SkillTreeView {
	s := &SkillTreeView{
		game:           game,
		focusLeft:      true,
		style:          backend.DefaultStyle(),
		unlockedStyle:  backend.DefaultStyle().Foreground(backend.ColorGreen),
		availableStyle: backend.DefaultStyle().Foreground(backend.ColorYellow),
		lockedStyle:    backend.DefaultStyle().Dim(true),
	}

	s.combatTree = widgets.NewTree(s.buildCombatTree())
	s.tradeTree = widgets.NewTree(s.buildTradeTree())
	s.combatTree.Focus()

	return s
}

func (s *SkillTreeView) buildCombatTree() *widgets.TreeNode {
	return &widgets.TreeNode{
		Label:    "Combat Skills",
		Expanded: true,
		Children: []*widgets.TreeNode{
			{
				Label:    s.formatSkill("Workout", s.game.workoutCount, 5, "$50"),
				Expanded: true,
				Children: []*widgets.TreeNode{
					{Label: s.formatSkill("Thick Skin", s.game.thickSkinCount, 5, "$75")},
				},
			},
			{Label: s.formatSkill("Track Practice", s.game.trackPracticeCount, 5, "$50")},
			{
				Label:    s.formatSkill("Hire Muscle", boolToInt(s.game.hasMuscle), 1, "$200"),
				Expanded: true,
				Children: []*widgets.TreeNode{
					{Label: s.formatSkill("Intimidation", boolToInt(s.game.hasIntimidation), 1, "$300")},
				},
			},
		},
	}
}

func (s *SkillTreeView) buildTradeTree() *widgets.TreeNode {
	return &widgets.TreeNode{
		Label:    "Trade Skills",
		Expanded: true,
		Children: []*widgets.TreeNode{
			{
				Label:    s.formatSkill("Backpack", s.game.backpackTier, 3, "$100+"),
				Expanded: true,
				Children: []*widgets.TreeNode{
					{Label: s.formatSkill("Secret Stash", boolToInt(s.game.hasStash), 1, "$150")},
				},
			},
			{Label: s.formatSkill("Bike", boolToInt(s.game.hasBike), 1, "$200")},
			{Label: s.formatSkill("Informant", boolToInt(s.game.hasInformant), 1, "$250")},
			{Label: s.formatSkill("Bank+", boolToInt(s.game.bankExpanded), 1, "$500")},
		},
	}
}

func (s *SkillTreeView) formatSkill(name string, current, max int, cost string) string {
	if current >= max {
		return "[+] " + name + " (MAX)"
	}
	if current > 0 {
		return fmt.Sprintf("[*] %s (%d/%d) %s", name, current, max, cost)
	}
	return fmt.Sprintf("[ ] %s %s", name, cost)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (s *SkillTreeView) Refresh() {
	s.combatTree.SetRoot(s.buildCombatTree())
	s.tradeTree.SetRoot(s.buildTradeTree())
}

func (s *SkillTreeView) Measure(c runtime.Constraints) runtime.Size {
	return c.MaxSize()
}

func (s *SkillTreeView) Layout(bounds runtime.Rect) {
	s.Component.Layout(bounds)

	halfWidth := bounds.Width / 2
	s.combatTree.Layout(runtime.Rect{X: bounds.X, Y: bounds.Y + 1, Width: halfWidth - 1, Height: bounds.Height - 1})
	s.tradeTree.Layout(runtime.Rect{X: bounds.X + halfWidth, Y: bounds.Y + 1, Width: bounds.Width - halfWidth, Height: bounds.Height - 1})
}

func (s *SkillTreeView) Render(ctx runtime.RenderContext) {
	s.Refresh()
	bounds := s.Bounds()

	// Headers
	ctx.Buffer.SetString(bounds.X, bounds.Y, "Combat", s.style.Bold(true))
	ctx.Buffer.SetString(bounds.X+bounds.Width/2, bounds.Y, "Trade", s.style.Bold(true))

	s.combatTree.Render(ctx)
	s.tradeTree.Render(ctx)
}

func (s *SkillTreeView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s.focusLeft {
		return s.combatTree.HandleMessage(msg)
	}
	return s.tradeTree.HandleMessage(msg)
}
