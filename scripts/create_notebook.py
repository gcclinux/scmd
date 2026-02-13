import json

# Read the JSONL file
dataset_path = '/home/ricardo/Programing/scmd/docs/markdown_finetune_scmd_dataset.jsonl'
with open(dataset_path, 'r') as f:
    dataset_content = [json.loads(line) for line in f]

# Create notebook structure
notebook = {
 "cells": [
  {
   "cell_type": "markdown",
   "metadata": {},
   "source": [
    "# Fine-tune Llama 3 for Markdown Training\n",
    "\n",
    "This notebook uses [Unsloth](https://github.com/unslothai/unsloth) to fine-tune a Llama 3 model on your custom Markdown dataset. Unsloth is optimized for free Colab GPUs (T4)."
   ]
  },
  {
   "cell_type": "code",
   "execution_count": None,
   "metadata": {},
   "outputs": [],
   "source": [
    "%%capture\n",
    "import torch\n",
    "major_version, minor_version = torch.cuda.get_device_capability()\n",
    "# Must install separately since Colab has torch 2.2.1, which breaks packages\n",
    "!pip install \"unsloth[colab-new] @ git+https://github.com/unslothai/unsloth.git\"\n",
    "if major_version >= 8:\n",
    "    # Use new 8-bit quantization for Ampere GPUs (A100, H100)\n",
    "    !pip install --no-deps packaging ninja einops flash-attn xformers trl peft accelerate bitsandbytes\n",
    "else:\n",
    "    # Use old quantization for Volta/Turing GPUs (T4, V100)\n",
    "    !pip install --no-deps xformers trl peft accelerate bitsandbytes\n",
    "pass"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": None,
   "metadata": {},
   "outputs": [],
   "source": [
    "from unsloth import FastLanguageModel\n",
    "import torch\n",
    "max_seq_length = 2048 # Choose any! We auto support RoPE Scaling internally!\n",
    "dtype = None # None for auto detection. Float16 for Tesla T4, V100, Bfloat16 for Ampere+\n",
    "load_in_4bit = True # Use 4bit quantization to reduce memory usage. Can be False.\n",
    "\n",
    "model, tokenizer = FastLanguageModel.from_pretrained(\n",
    "    model_name = \"unsloth/llama-3-8b-bnb-4bit\", # Llama-3 8b 4bit quantized\n",
    "    max_seq_length = max_seq_length,\n",
    "    dtype = dtype,\n",
    "    load_in_4bit = load_in_4bit,\n",
    ")"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": None,
   "metadata": {},
   "outputs": [],
   "source": [
    "model = FastLanguageModel.get_peft_model(\n",
    "    model,\n",
    "    r = 16, # Choose any number > 0 ! Suggested 8, 16, 32, 64, 128\n",
    "    target_modules = [\"q_proj\", \"k_proj\", \"v_proj\", \"o_proj\",\n",
    "                      \"gate_proj\", \"up_proj\", \"down_proj\",],\n",
    "    lora_alpha = 16,\n",
    "    lora_dropout = 0, # Supports any, but = 0 is optimized\n",
    "    bias = \"none\",    # Supports any, but = \"none\" is optimized\n",
    "    use_gradient_checkpointing = \"unsloth\", # 4x longer context window + lower VRAM usage.\n",
    "    random_state = 3407,\n",
    "    use_rslora = False,  # We support rank stabilized LoRA\n",
    "    loftq_config = None, # And LoftQ\n",
    ")"
   ]
  },
  {
   "cell_type": "markdown",
   "metadata": {},
   "source": [
    "### Load Dataset\n",
    "We will load your custom dataset directly here."
   ]
  },
  {
   "cell_type": "code",
   "execution_count": None,
   "metadata": {},
   "outputs": [],
   "source": [
    "from datasets import Dataset\n",
    "\n",
    "# Your Custom Dataset injected directly\n",
    "custom_data = " + json.dumps(dataset_content, indent=4) + "\n",
    "\n",
    "# Convert to HuggingFace Dataset format\n",
    "dataset = Dataset.from_list(custom_data)\n",
    "\n",
    "alpaca_prompt = \"\"\"Below is an instruction that describes a task, paired with an input that provides further context. Write a response that appropriately completes the request.\n",
    "\n",
    "### Instruction:\n",
    "{}\n",
    "\n",
    "### Input:\n",
    "{}\n",
    "\n",
    "### Response:\n",
    "{}\"\"\"\n",
    "\n",
    "EOS_TOKEN = tokenizer.eos_token # Must add EOS_TOKEN\n",
    "def formatting_prompts_func(examples):\n",
    "    instructions = examples[\"instruction\"]\n",
    "    inputs       = examples[\"input\"]\n",
    "    outputs      = examples[\"output\"]\n",
    "    texts = []\n",
    "    for instruction, input, output in zip(instructions, inputs, outputs):\n",
    "        # Must add EOS_TOKEN, otherwise your generation will go on forever!\n",
    "        text = alpaca_prompt.format(instruction, input, output) + EOS_TOKEN\n",
    "        texts.append(text)\n",
    "    return { \"text\" : texts, }\n",
    "pass\n",
    "\n",
    "# Apply formatting\n",
    "dataset = dataset.map(formatting_prompts_func, batched = True)"
   ]
  },
  {
   "cell_type": "markdown",
   "metadata": {},
   "source": [
    "### Train the Model"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": None,
   "metadata": {},
   "outputs": [],
   "source": [
    "from trl import SFTTrainer\n",
    "from transformers import TrainingArguments\n",
    "\n",
    "trainer = SFTTrainer(\n",
    "    model = model,\n",
    "    tokenizer = tokenizer,\n",
    "    train_dataset = dataset,\n",
    "    dataset_text_field = \"text\",\n",
    "    max_seq_length = max_seq_length,\n",
    "    dataset_num_proc = 2,\n",
    "    packing = False, # Can make training 5x faster for short sequences.\n",
    "    args = TrainingArguments(\n",
    "        per_device_train_batch_size = 2,\n",
    "        gradient_accumulation_steps = 4,\n",
    "        warmup_steps = 5,\n",
    "        max_steps = 60, # Adjusted for small dataset (1 epoch approx)\n",
    "        learning_rate = 2e-4,\n",
    "        fp16 = not torch.cuda.is_bf16_supported(),\n",
    "        bf16 = torch.cuda.is_bf16_supported(),\n",
    "        logging_steps = 1,\n",
    "        optim = \"adamw_8bit\",\n",
    "        weight_decay = 0.01,\n",
    "        lr_scheduler_type = \"linear\",\n",
    "        seed = 3407,\n",
    "        output_dir = \"outputs\",\n",
    "    ),\n",
    ")"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": None,
   "metadata": {},
   "outputs": [],
   "source": [
    "trainer_stats = trainer.train()"
   ]
  },
  {
   "cell_type": "markdown",
   "metadata": {},
   "source": [
    "### Inference (Test It)"
   ]
  },
  {
   "cell_type": "code",
   "execution_count": None,
   "metadata": {},
   "outputs": [],
   "source": [
    "FastLanguageModel.for_inference(model) # Enable native 2x faster inference\n",
    "inputs = tokenizer(\n",
    "[\n",
    "    alpaca_prompt.format(\n",
    "        \"How do I make a table in Markdown?\", # Instruction\n",
    "        \"\", # Input\n",
    "        \"\", # Output - leave this blank for generation!\n",
    "    )\n",
    "], return_tensors = \"pt\").to(\"cuda\")\n",
    "\n",
    "outputs = model.generate(**inputs, max_new_tokens = 128, use_cache = True)\n",
    "tokenizer.batch_decode(outputs)"
   ]
  },
  {
   "cell_type": "markdown",
   "metadata": {},
   "source": [
    "### Save to GGUF\n",
    "This will save the model in a format you can use with Ollama."
   ]
  },
  {
   "cell_type": "code",
   "execution_count": None,
   "metadata": {},
   "outputs": [],
   "source": [
    "# Save to 8bit Q8_0\n",
    "model.save_pretrained_gguf(\"model_gguf\", tokenizer, quantization_method = \"q8_0\")\n",
    "# model.save_pretrained_gguf(\"model_gguf\", tokenizer, quantization_method = \"f16\")"
   ]
  }
 ],
 "metadata": {
  "kernelspec": {
   "display_name": "Python 3",
   "language": "python",
   "name": "python3"
  },
  "language_info": {
   "codemirror_mode": {
    "name": "ipython",
    "version": 3
   },
   "file_extension": ".py",
   "mimetype": "text/x-python",
   "name": "python",
   "nbconvert_exporter": "python",
   "pygments_lexer": "ipython3",
   "version": "3.10.12"
  }
 },
 "nbformat": 4,
 "nbformat_minor": 2
}

# Write notebook file
with open('/home/ricardo/Programing/scmd/docs/Markdown_Llama3_Finetune.ipynb', 'w') as f:
    json.dump(notebook, f, indent=2)
