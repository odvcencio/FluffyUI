# Migrating from tview

This guide highlights common mappings between tview and FluffyUI.

## Core Concepts

| tview | FluffyUI |
|------|----------|
| Application | runtime.App / fluffy.NewApp |
| Primitive | runtime.Widget |
| SetRoot | App.SetRoot |
| Draw | services.Invalidate() |

## Widgets

| tview | FluffyUI |
|------|----------|
| TextView | widgets.Text / widgets.TextArea |
| InputField | widgets.Input |
| List | widgets.List |
| Table | widgets.Table / widgets.DataGrid |
| TreeView | widgets.Tree |
| Modal | widgets.Dialog |

## Layout

| tview | FluffyUI |
|------|----------|
| Flex | widgets.Flex |
| Grid | widgets.Grid |
| Pages | widgets.Stack |

## Notes

- FluffyUI uses a measure/layout/render pipeline; avoid direct drawing.
- Reactive signals (`state.Signal`) replace manual state tracking.
