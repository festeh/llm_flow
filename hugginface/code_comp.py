from huggingface_hub import InferenceClient
import os 


token = os.environ["HF_API_TOKEN"]
client = InferenceClient(model="codellama/CodeLlama-7b-hf", token=token)

prompt_prefix = """
local ui = require("llm_flow.ui")
local uv = vim.uv
local l = require("llm_flow.logger")


local kDebounce = 100 -- ms

local M = {
  suggestion = nil,
  timer = nil,
  client = nil,
  req_id = nil,
}

local function stop_timer_and_cancel()
  ui.clear()
  M.req_id = nil
  if M.timer then
    M.timer:stop()
    M.timer:close()
    M.timer = nil
  end
  local client = M.client
  if client then
    for req_id, req in pairs(client.requests) do
      if req.type == "pending" then
        l.log(req_id, "pending")
        client.notify('cancel_predict_editor', { id = req_id })
      end
    end
    l.log("\n-\n")
  end
end

function M.find_lsp_client()
  local clients = vim.lsp.get_clients({ bufnr = 0 })
  for _, client in pairs(clients) do
    if client.name == "llm-flow" then
      M.client = client
      return client
    end
  end
end

local function on_predict_complete(err, result, line, pos)
  if err then
    vim.notify("Prediction failed: " .. err.message, vim.log.levels.ERROR)
    return
  end

  -- Return if not in insert mode
  if vim.api.nvim_get_mode().mode ~= "i" then
    return
  end

  if M.req_id ~= result.id then
    l.log("rejected", "expected", M.req_id, "got", result.id)
    return
  end

  local content = result.content
  M.suggestion = {
    content = content,
    line = line,
    pos = pos
  }
  local content_lines = vim.split(content, "\n", { plain = true })
  local truncated_content = { content_lines[1] }
  local bufnr = vim.api.nvim_get_current_buf()
  local buffer_lines = vim.api.nvim_buf_get_lines(bufnr, line, line + #content_lines, false)
  for i = 2, #content_lines do
    local trimmed_content = vim.trim(content_lines[i] or "")
    local trimmed_buffer = vim.trim(buffer_lines[i] or "")
    if trimmed_content == trimmed_buffer then
      break
    end
    table.insert(truncated_content, content_lines[i])
  end
  local final_content = table.concat(truncated_content, "\n")
  ui.set_text(line, pos, final_content)
  l.log(result.id, "completed")
  return result
end

--- @param params table The parameters for the prediction
function M.predict_editor(params)
  local client = M.find_lsp_client()

  if not client then
    return
  end

  params = params or {}
  local bufnr = vim.api.nvim_get_current_buf()
  local cursor = vim.api.nvim_win_get_cursor(0)

  local line = cursor[1] - 1
  local pos = cursor[2]

  local request_params = vim.tbl_extend("force", params, {
    provider = "huggingface",
    model = "codellama/CodeLlama-13b-hf",
    uri = vim.uri_from_bufnr(bufnr),
    line = line,
    pos = pos
  })
  local status, req_id = client.request("predict_editor", request_params, function(err, result)
    on_predict_complete(err, result, line, pos)
  end)
  M.req_id = req_id
  l.log("Sent request", req_id)
""".strip()
prompt_suffix = ""

prompt = f"<PRE> {prompt_prefix} <SUF>{prompt_suffix} <MID>"

infilled = client.text_generation(prompt, max_new_tokens=10)
print("INFILLED", infilled)
infilled = infilled.rstrip(" <EOT>")
