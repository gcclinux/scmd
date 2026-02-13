import json
import os

file_path = '/home/ricardo/Programing/scmd/docs/markdown_finetune_scmd_dataset.jsonl'
temp_path = file_path + '.tmp'

updated_count = 0

with open(file_path, 'r') as infile, open(temp_path, 'w') as outfile:
    for line in infile:
        if not line.strip():
            continue
            
        try:
            data = json.loads(line)
            instruction = data.get("instruction", "").strip()
            
            # Check if "markdown" is already mentioned (case-insensitive)
            if "markdown" not in instruction.lower():
                # Logic to append "in Markdown" appropriately
                if instruction.endswith("?"):
                    instruction = instruction[:-1] + " in Markdown?"
                elif instruction.endswith("."):
                    instruction = instruction[:-1] + " in Markdown."
                else:
                    instruction = instruction + " in Markdown"
                
                data["instruction"] = instruction
                updated_count += 1
            
            outfile.write(json.dumps(data) + '\n')
            
        except json.JSONDecodeError:
            print(f"Skipping invalid JSON line: {line}")

# Replace the original file with the updated one
os.replace(temp_path, file_path)

print(f"Successfully updated {updated_count} instructions in {file_path}")
