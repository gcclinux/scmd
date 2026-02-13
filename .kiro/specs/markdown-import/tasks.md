# Implementation Plan: Markdown Import

## Overview

Implement markdown document import for scmd with two entry points (`/import` and `--import`), vector embedding generation, and terminal markdown rendering for search results. All new logic lives in two new files (`importmd.go`, `rendermd.go`) with minimal changes to existing files.

## Tasks

- [x] 1. Implement markdown import core logic
  - [x] 1.1 Create `importmd.go` with `ImportMarkdown`, `extractTitle`, and `isMarkdownFile` functions
    - `isMarkdownFile` checks for `.md` extension (case-insensitive)
    - `extractTitle` scans for first `# ` line, falls back to filename without extension
    - `ImportMarkdown` orchestrates: validate extension → read file → check empty → extract title → check duplicate → call `AddCommand`
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 2.1, 2.4, 5.1, 5.2, 5.3_

  - [ ]* 1.2 Write property tests for `isMarkdownFile` and `extractTitle`
    - **Property 1: Non-markdown extensions are rejected**
    - **Validates: Requirements 1.3**
    - **Property 2: Title extraction from heading**
    - **Validates: Requirements 1.4**
    - **Property 3: Title fallback to filename**
    - **Validates: Requirements 1.5**

- [x] 2. Implement markdown terminal renderer
  - [x] 2.1 Create `rendermd.go` with `RenderMarkdown` and `isMarkdownContent` functions
    - `isMarkdownContent` detects fenced code blocks, headers, and markdown links
    - `RenderMarkdown` applies ANSI styling to headers and renders code blocks with visible delimiters and language labels
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [ ]* 2.2 Write property tests for `isMarkdownContent` and `RenderMarkdown`
    - **Property 4: Markdown detection accuracy**
    - **Validates: Requirements 4.3, 4.4**
    - **Property 5: Code block rendering preserves delimiters**
    - **Validates: Requirements 4.1**
    - **Property 6: Header rendering applies ANSI styling**
    - **Validates: Requirements 4.2**

- [x] 3. Checkpoint
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Wire import into interactive CLI
  - [x] 4.1 Add `/import` command to `handleSlashCommand` in `interactive.go`
    - Add case for `/import` that calls a new `handleImportCommand(args)` function
    - `handleImportCommand` initializes Gemini, Ollama, and DB, calls `ImportMarkdown`, prints success/error
    - Add `/import` to `printInteractiveHelp`
    - _Requirements: 1.1, 1.2, 1.3, 1.6, 1.7_

  - [x] 4.2 Add `--import` flag handling to `main.go`
    - Add `--import` case in the `count == 3` branch
    - Initialize embedding providers and DB, call `ImportMarkdown`, print result, handle errors with stderr and exit code 1
    - _Requirements: 2.1, 2.2, 2.3_

  - [x] 4.3 Add `--import` usage to `helpmenu.go`
    - Add help text for `--import [filepath]` to the help menu
    - _Requirements: 2.1_

- [x] 5. Integrate markdown rendering into search results
  - [x] 5.1 Update `performInteractiveSearch` in `interactive.go` to use markdown renderer
    - Before displaying `result.Key` or `result.Data`, check `isMarkdownContent`
    - If markdown, pass through `RenderMarkdown` instead of the existing plain/code display logic
    - Preserve existing behavior for non-markdown content
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [ ]* 5.2 Write unit tests for content preservation round-trip
    - **Property 7: Content preservation round-trip**
    - **Validates: Requirements 5.3**

- [x] 6. Final checkpoint
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- The `ImportMarkdown` function reuses the existing `AddCommand` for storage and embedding generation, keeping changes minimal
- No database schema changes are needed — documents use the same `CommandRecord` table
- Property tests use the `gopter` library with minimum 100 iterations per property
