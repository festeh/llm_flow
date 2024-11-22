from huggingface_hub import InferenceClient
import os 


token = os.environ["HF_API_TOKEN"]
client = InferenceClient(model="codellama/CodeLlama-7b-hf", token=token)

prompt_prefix = 'def a_plus_b(a, '
prompt_suffix = ""

prompt = f"<PRE> {prompt_prefix} <SUF>{prompt_suffix} <MID>"

infilled = client.text_generation(prompt, max_new_tokens=10)
infilled = infilled.rstrip(" <EOT>")
print(f"{prompt_prefix}{infilled}{prompt_suffix}")
