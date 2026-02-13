# Requirements Document

## Introduction

This feature adds markdown document import capability to the scmd CLI tool. Currently, scmd stores individual commands with descriptions. This feature extends scmd to import entire markdown documents, parse their structure, generate vector embeddings for semantic search, and store them in PostgreSQL. It also enhances the interactive CLI to render search results with proper markdown formatting (code blocks, headers, etc.) instead of plain text.

## Glossary

- **SCMD**: The command-line tool for saving, searching, and managing commands with AI-powered semantic search.
- **Interactive_CLI**: The interactive mode of SCMD, started via `--interactive`, `-i`, or `--cli` flags, where users issue slash commands.
- **Markdown_Importer**: The component responsible for reading a markdown file from disk, parsing its content, and preparing it for storage.
- **Document_Record**: A database entry representing an imported markdown document, stored with its title as the key and full markdown content as the data.
- **Embedding_Provider**: An external service (Ollama or Gemini) used to generate vector embeddings from text content.
- **Markdown_Renderer**: The component responsible for formatting markdown content with visual styling (code block highlighting, headers, etc.) in the terminal output.

## Requirements

### Requirement 1: Import Markdown via Interactive CLI

**User Story:** As a user, I want to import a markdown document from the interactive CLI using `/import <path>`, so that I can quickly add documentation to my searchable command database without leaving the interactive session.

#### Acceptance Criteria

1. WHEN a user issues `/import <path>` in the Interactive_CLI, THE Markdown_Importer SHALL read the file at the specified path and store it as a Document_Record in the database.
2. WHEN the specified file path does not exist, THE Markdown_Importer SHALL display a descriptive error message indicating the file was not found.
3. WHEN the specified file is not a valid markdown file (not `.md` extension), THE Markdown_Importer SHALL display an error message indicating only markdown files are supported.
4. WHEN the file is read successfully, THE Markdown_Importer SHALL extract the first top-level heading (`# Title`) as the document key and use the full markdown content as the data.
5. WHEN the markdown file contains no top-level heading, THE Markdown_Importer SHALL use the filename (without extension) as the document key.
6. WHEN a document with the same key already exists in the database, THE Markdown_Importer SHALL display an error message indicating the document already exists.
7. WHEN the import completes successfully, THE Interactive_CLI SHALL display a confirmation message with the document title and a success indicator.

### Requirement 2: Import Markdown via Command-Line Flag

**User Story:** As a user, I want to import a markdown document from the command line using `--import <path>`, so that I can add documentation to my database from scripts or one-off commands.

#### Acceptance Criteria

1. WHEN a user runs `scmd --import <path>`, THE Markdown_Importer SHALL read the file at the specified path and store it as a Document_Record in the database.
2. WHEN the specified file path does not exist, THE Markdown_Importer SHALL print an error message to stderr and exit with a non-zero exit code.
3. WHEN the specified file is not a valid markdown file (not `.md` extension), THE Markdown_Importer SHALL print an error message to stderr and exit with a non-zero exit code.
4. THE Markdown_Importer SHALL use the same parsing and storage logic for both the `--import` flag and the `/import` interactive command.

### Requirement 3: Generate Embeddings for Imported Documents

**User Story:** As a user, I want imported markdown documents to have vector embeddings generated automatically, so that I can find them through semantic AI-powered search.

#### Acceptance Criteria

1. WHEN a markdown document is imported, THE Markdown_Importer SHALL generate a vector embedding from the combined document key and content using the available Embedding_Provider.
2. WHEN Ollama is available, THE Markdown_Importer SHALL use Ollama as the primary Embedding_Provider for generating the vector embedding.
3. WHEN Ollama is unavailable and Gemini is available, THE Markdown_Importer SHALL fall back to Gemini as the Embedding_Provider.
4. IF no Embedding_Provider is available, THEN THE Markdown_Importer SHALL store the Document_Record without an embedding and log a warning.
5. THE Markdown_Importer SHALL store the generated embedding alongside the Document_Record in the same PostgreSQL table used for commands.

### Requirement 4: Markdown-Formatted Search Results in Interactive CLI

**User Story:** As a user, I want search results in the interactive CLI to render markdown content with proper formatting (code blocks, headers), so that I can read imported documentation clearly in the terminal.

#### Acceptance Criteria

1. WHEN displaying a search result that contains markdown content, THE Markdown_Renderer SHALL render fenced code blocks with visible delimiters and syntax highlighting labels.
2. WHEN displaying a search result that contains markdown headers, THE Markdown_Renderer SHALL render headers with distinct visual styling (e.g., bold or color).
3. WHEN displaying a search result that contains plain command text (non-markdown), THE Markdown_Renderer SHALL display the result using the existing formatting logic.
4. WHEN displaying a search result, THE Markdown_Renderer SHALL detect whether the content is markdown or a plain command and choose the appropriate rendering path.

### Requirement 5: Markdown File Parsing

**User Story:** As a developer, I want a reliable markdown file parser, so that the import feature correctly reads and processes markdown documents.

#### Acceptance Criteria

1. THE Markdown_Importer SHALL read the entire file content as a UTF-8 encoded string.
2. WHEN the file content is empty, THE Markdown_Importer SHALL display an error message indicating the file is empty and not import it.
3. THE Markdown_Importer SHALL preserve the original markdown formatting (code blocks, lists, headers, links) in the stored data field.
