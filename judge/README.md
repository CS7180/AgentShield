# Judge Service

Standalone HTTP service that evaluates attack outputs and returns severity/confidence.

It supports two evaluation modes:

- `rule` (default): rule-based local scoring
- `openai_compat`: calls an OpenAI-compatible `chat/completions` endpoint and expects JSON output

## Endpoints

- `GET /health`
- `POST /evaluate-batch`

## Run

```bash
cd judge
make run
```

Environment:

- `JUDGE_PORT` (default `8091`)
- `JUDGE_EVAL_MODE` (`rule` or `openai_compat`, default `rule`)
- `JUDGE_LLM_BASE_URL` (required for `openai_compat`)
- `JUDGE_LLM_API_KEY` (required for `openai_compat`)
- `JUDGE_LLM_MODEL` (required for `openai_compat`)
