# Migrating from termui

This guide highlights common mappings between termui and FluffyUI.

## Core Concepts

| termui | FluffyUI |
|------|----------|
| ui.Render | services.Invalidate() |
| ui.PollEvents | runtime.App loop |
| Drawable | runtime.Widget |

## Widgets

| termui | FluffyUI |
|------|----------|
| List | widgets.List / widgets.VirtualList |
| Table | widgets.Table / widgets.DataGrid |
| Paragraph | widgets.Text / widgets.TextArea |
| Gauge | widgets.Gauge / widgets.AnimatedGauge |
| Sparkline | widgets.LineChart |

## Layout

| termui | FluffyUI |
|------|----------|
| Grid | widgets.Grid |
| Rows/Cols | widgets.Flex |

## Notes

- Prefer options pattern for widget construction.
- Use the simulation backend for deterministic testing.
