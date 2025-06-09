# AI Dojo

AI Dojo is a collection of experiments and pipelines for extracting structured data from unstructured sources using modern AI and open-source tools. This repository demonstrates how to bridge the gap between raw, messy data (like PDFs or free-form text) and structured, actionable information.

## Features

- **PDF Text Extraction**: Extract human-readable text from PDF documents using open-source Go libraries.
- **Document Classification**: Automatically classify documents (e.g., Resume, Cover Letter) using AI models.
- **Feature Extraction**: Pull out structured features such as contact info, skills, and experience from unstructured documents.
- **SQL Pipelines**: Convert unstructured data into structured outputs for downstream analytics or processing.
- **FastMCP Integration**: Python-based microservice for rapid prototyping and serving AI-powered tools.

## Repository Structure

```
ai-dojo/
├── 4-data-extraction-unstructured/      # Go code for PDF and unstructured data extraction
├── sql-pipeline/                       # Structured output pipelines and processors
│   └── 2-structured-output/
│       └── unstructured-processor/
├── mcp/                                # Python FastMCP microservice
├── README.md
```

## Getting Started

### Go Projects

1. **Install Go dependencies**  
   In each Go subfolder:
   ```bash
   go mod tidy
   ```

2. **Run Go code**
   ```bash
   go run main.go
   ```

### Python (FastMCP)

1. **Install Poetry**  
   Follow the [Poetry installation guide](https://python-poetry.org/docs/#installation).

2. **Install dependencies**
   ```bash
   cd mcp
   poetry install
   ```

3. **Run the FastMCP server**
   ```bash
   poetry run python main.py
   ```
   The server runs by default on [http://localhost:8000](http://localhost:8000).

## Development

- **Go**: Run tests with  
  ```bash
  go test ./...
  ```
- **Python**: Run tests with  
  ```bash
  poetry run pytest
  ```

## Environment Variables

Some features require API keys (e.g., OpenAI). Set them in your shell or a `.env` file:
```
OPENAI_API_KEY=your-api-key
```

## Contributing

Pull requests and issues are welcome! Please open an issue to discuss your ideas or report bugs.

## License

This repository is licensed under the MIT License.

---