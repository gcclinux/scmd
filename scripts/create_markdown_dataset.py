import json
import re

def create_dataset():
    dataset = []

    # 1. Manual Examples (Augmenting with more details)
    manual_examples = [
        {
            "instruction": "How do I create a level 1 header in Markdown?",
            "input": "",
            "output": "# Heading Level 1"
        },
        {
            "instruction": "Create a level 2 header with the text 'Introduction'.",
            "input": "",
            "output": "## Introduction"
        },
        {
            "instruction": "Show me an example of bold text in Markdown using asterisks.",
            "input": "",
            "output": "**Bold Text**"
        },
        {
            "instruction": "How do I italicize text using underscores?",
            "input": "",
            "output": "_Italic Text_"
        },
        {
            "instruction": "Create an unordered list with three items: Apple, Banana, Orange.",
            "input": "",
            "output": "- Apple\n- Banana\n- Orange"
        },
        {
            "instruction": "How do I create a link to Google?",
            "input": "",
            "output": "[Google](https://www.google.com)"
        },
        {
            "instruction": "Insert an image with alt text 'Logo' and source 'logo.png'.",
            "input": "",
            "output": "![Logo](logo.png)"
        },
        {
            "instruction": "Create a code block for Python code that prints 'Hello'.",
            "input": "",
            "output": "```python\nprint('Hello')\n```"
        },
        {
            "instruction": "How do I make a blockquote?",
            "input": "",
            "output": "> This is a blockquote."
        },
        {
            "instruction": "Create a table with two columns: Name and Age.",
            "input": "",
            "output": "| Name | Age |\n| --- | --- |\n| Alice | 30 |\n| Bob | 25 |"
        },
        {
            "instruction": "How do I add a horizontal rule?",
            "input": "",
            "output": "---"
        },
        {
            "instruction": "Show me how to do a strikethrough.",
            "input": "",
            "output": "~~Strikethrough~~"
        },
        {
            "instruction": "Create a task list with one done item and one todo item.",
            "input": "",
            "output": "- [x] Done item\n- [ ] Todo item"
        },
        {
            "instruction": "How do I escape a special character like an asterisk?",
            "input": "",
            "output": "\\*"
        },
        {
            "instruction": "Create a link that has a title attribute.",
            "input": "",
            "output": "[Link Text](https://example.com \"Title Text\")"
        },
        {
            "instruction": "How do I create a definition list?",
            "input": "",
            "output": "Term\n: Definition"
        },
        {
            "instruction": "Show me how to footnote a sentence.",
            "input": "",
            "output": "This is a sentence.[^1]\n\n[^1]: This is the footnote."
        },
        {
            "instruction": "How do I write a superscript?",
            "input": "",
            "output": "X^2^"
        },
        {
            "instruction": "How do I write a subscript?",
            "input": "",
            "output": "H~2~O"
        },
        {
             "instruction": "Create a collapse/expand details section in Markdown (using HTML).",
             "input": "",
             "output": "<details>\n<summary>Click to expand</summary>\n\nHidden content goes here.\n</details>"
        },
        {
            "instruction": "How do I highlight text (if supported)?",
            "input": "",
            "output": "==Highlighted Text=="
        },
        {
            "instruction": "Create a nested list structure.",
            "input": "",
            "output": "1. First Item\n   - Nested Item 1\n   - Nested Item 2\n2. Second Item"
        },
        {
             "instruction": "How can I comment out text in Markdown?",
             "input": "",
             "output": "<!-- This is a comment and will not be rendered -->"
        },
        {
            "instruction": "Create a mermaid class diagram example.",
            "input": "",
            "output": "```mermaid\nclassDiagram\n    Animal <|-- Duck\n    Animal <|-- Fish\n    Animal <|-- Zebra\n    Animal : +int age\n    Animal : +String gender\n    class Duck{\n      +String beakColor\n      +swim()\n      +quack()\n    }\n    class Fish{\n      -int sizeInFeet\n      -canEat()\n    }\n```"
        },
        {
           "instruction": "What is the syntax for a math block (KaTeX/MathJax)?",
           "input": "",
           "output": "$$ E = mc^2 $$" 
        }
    ]
    dataset.extend(manual_examples)

    # 2. Parse existing file for "Context-based" learning
    # We will read the file and create broader "Explain X" style instructions
    
    file_path = '/home/ricardo/Programing/scmd/docs/MARKDOWN_TRAINING_DATA.md'
    with open(file_path, 'r') as f:
        content = f.read()

    # Extract sections by ## Headers
    sections = re.split(r'^##\s+', content, flags=re.MULTILINE)[1:] # Skip preamble
    
    for section in sections:
        lines = section.split('\n')
        title = lines[0].strip()
        # Remove numbering if present (e.g., "1. Basic Syntax" -> "Basic Syntax")
        info_title = re.sub(r'^\d+\.\s+', '', title) 
        body = '\n'.join(lines[1:]).strip()
        
        if not body:
            continue

        # Create a general instruction for the section
        dataset.append({
            "instruction": f"Explain Markdown syntax for {info_title}.",
            "input": "",
            "output": body
        })

        # Create a "Give me examples" instruction
        dataset.append({
            "instruction": f"Provide examples of {info_title} in Markdown.",
            "input": "",
            "output": body
        })

    # Add a full document generation task
    dataset.append({
        "instruction": "Generate a comprehensive Markdown cheat sheet.",
        "input": "",
        "output": content
    })

    # 3. Write to JSONL
    output_path = '/home/ricardo/Programing/scmd/docs/markdown_finetune_dataset.jsonl'
    with open(output_path, 'w') as f:
        for entry in dataset:
            f.write(json.dumps(entry) + '\n')

    print(f"Dataset created at {output_path} with {len(dataset)} examples.")

if __name__ == "__main__":
    create_dataset()
