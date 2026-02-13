import json
import subprocess
import os
import sys

# Configuration
DATASET_PATH = '/home/ricardo/Programing/scmd/docs/markdown_finetune_scmd_dataset.jsonl'
# Assuming scmd binary is in the project root based on file structure
SCMD_BINARY = '/home/ricardo/Programing/scmd/scmd'

def import_data():
    if not os.path.exists(DATASET_PATH):
        print(f"Error: Dataset not found at {DATASET_PATH}")
        sys.exit(1)
        
    # Check if binary exists at absolute path, otherwise try 'scmd' from PATH
    executable = SCMD_BINARY
    if not os.path.exists(executable):
        executable = 'scmd'
        # verify if it's in path
        if subprocess.call(['which', 'scmd'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL) != 0:
             print(f"Error: 'scmd' binary not found at {SCMD_BINARY} and not in PATH.")
             sys.exit(1)

    print(f"Using executable: {executable}")
    print(f"Reading from: {DATASET_PATH}")
    print("-" * 50)

    success_count = 0
    error_count = 0

    with open(DATASET_PATH, 'r') as f:
        for i, line in enumerate(f):
            line = line.strip()
            if not line:
                continue

            try:
                data = json.loads(line)
                
                # Mapping:
                # instruction (Description) -> arg 2
                # output (Command/Content)  -> arg 1
                
                description = data.get('instruction', '').strip()
                command_content = data.get('output', '').strip()

                if not description or not command_content:
                    print(f"Line {i+1}: Skipping incomplete entry")
                    continue

                # Construct command: scmd --save [command] [description]
                # subprocess handles quote escaping for us
                args = [executable, '--save', command_content, description]
                
                # print(f"Processing: {description[:40]}...")
                
                result = subprocess.run(
                    args,
                    capture_output=True,
                    text=True
                )

                if result.returncode == 0:
                    # print(f"Success: {description[:30]}...")
                    success_count += 1
                else:
                    print(f"Failed to save line {i+1}: {description}")
                    print(f"Error: {result.stderr.strip()}")
                    print(f"Output: {result.stdout.strip()}")
                    error_count += 1
                    
            except json.JSONDecodeError:
                print(f"Line {i+1}: Invalid JSON")
                error_count += 1
            except Exception as e:
                print(f"Line {i+1}: Unexpected error: {str(e)}")
                error_count += 1

    print("-" * 50)
    print(f"Import completed.")
    print(f"Successfully added: {success_count}")
    print(f"Errors: {error_count}")

if __name__ == "__main__":
    import_data()
